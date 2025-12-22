package util

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var errIllegalPath = errors.New("illegal file path in archive")

// sanitizePath checks if name is safe to join with dest.
func sanitizePath(dest, name string) (string, error) {
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return "", err
	}
	target := filepath.Join(absDest, name)
	cleanTarget := filepath.Clean(target)
	cleanDest := filepath.Clean(absDest)

	if !strings.HasPrefix(cleanTarget, cleanDest+string(os.PathSeparator)) && cleanTarget != cleanDest {
		return "", fmt.Errorf("%w: %s", errIllegalPath, name)
	}
	return cleanTarget, nil
}

// stripPath strips leading path components and returns the remaining path.
func stripPath(name string, stripComponents int) string {
	name = filepath.ToSlash(name)
	parts := strings.Split(name, "/")
	if len(parts) <= stripComponents {
		return ""
	}
	return filepath.Join(parts[stripComponents:]...)
}

// ExtractTarGz extracts a .tar.gz archive to dest, stripping stripComponents directories.
// Only regular files and directories are extracted; symlinks are skipped.
func ExtractTarGz(src, dest string, stripComponents int) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		relPath := stripPath(header.Name, stripComponents)
		if relPath == "" {
			continue
		}

		target, err := sanitizePath(dest, relPath)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := extractTarFile(target, header.Mode, tr); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractTarFile(target string, mode int64, r io.Reader) (err error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(mode))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(f, r)
	return err
}

// ExtractGz extracts a single .gz file to dest.
func ExtractGz(src, dest string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, gzr)
	return err
}

// ExtractZip extracts a .zip archive to dest, stripping stripComponents directories.
func ExtractZip(src, dest string, stripComponents int) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		if err := extractZipFile(f, absDest, stripComponents); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(f *zip.File, dest string, stripComponents int) (err error) {
	relPath := stripPath(f.Name, stripComponents)
	if relPath == "" {
		return nil
	}

	target, err := sanitizePath(dest, relPath)
	if err != nil {
		return err
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, rc)
	return err
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	return err
}