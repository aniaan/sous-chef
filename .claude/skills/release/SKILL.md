---
name: release
description: Bump the sous-chef plugin version in metadata.lua, commit, push, and create a GitHub release with `gh`. Use this skill whenever the user wants to release a new version, cut a release, ship a version, publish to GitHub releases, tag a release, or bump the sous-chef plugin version — even if they only say "release" or "发个新版本" without specifying a version number. Project-specific to sous-chef; binaries are built and attached by CI, not by this skill.
---

# Release sous-chef

End-to-end workflow for cutting a new sous-chef plugin release: bump the version, commit, push, and create the GitHub release. Binaries are produced by CI on tag push — do NOT build or upload them from here.

## When to use this

Trigger phrases include "release a new version", "cut v0.0.X", "ship the next version", "升级版本并发布", "发个 release", "publish to github releases", or any request to bump the plugin version.

`metadata.lua`'s `version` field is the source of truth (per `AGENTS.md`'s "Releasing a New Version" rule); that's the file that gets bumped.

## Workflow

Walk through these steps in order. Don't skip steps silently — if something looks off (diverged branch, dirty tree, existing tag), pause and surface it to the user.

### 1. Sync with origin

Run `git status` and `git fetch && git pull --ff-only`.

The bump commit must sit on top of the latest `origin/main`. Things to watch for:
- **Diverged branch** → stop, ask the user how to resolve. Don't auto-rebase or force-push.
- **Uncommitted changes unrelated to the release** → ask whether to fold them in or stash. The user might have in-flight work they don't want bundled into the version commit.
- **Local commits not yet on origin** → that's fine; they'll go up with the bump.

### 2. Read the current version

Read `metadata.lua` and grab the `version = "X.Y.Z"` line. Also run `git tag --sort=-v:refname | head -3` and `gh release list --limit 3` to cross-check that origin's latest release matches what `metadata.lua` claims — if they disagree (e.g. metadata says 0.0.7 but a v0.0.8 release exists), the local branch is probably out of date and step 1 should have caught it. Re-pull or surface the mismatch.

### 3. Propose the next version

Default proposal: **patch +1** (e.g. `0.0.8 → 0.0.9`). Show it as a one-line proposal and let the user confirm, override (minor/major), or type a specific version. Wait for confirmation.

Don't propose anything pre-1.0 fancier than patch bumps unless the user asks — sous-chef has stayed on patch increments throughout its history.

### 4. Update version files

Bump the version in **both** files:

- `metadata.lua` — change only the `version = "..."` value; leave `name`, `description`, `author`, `repo` untouched.
- `Makefile` — change only the `VERSION=X.Y.Z` line at the top; this constant is baked into the binary via `-ldflags "-X main.Version=$(VERSION)"`, so keeping it in sync ensures `sous-chef --version` (or equivalent) reports the right number.

Edit only — do **not** run `make build` or `make release`. Building is CI's job.

### 5. Commit

Look at recent commit style first:
```bash
git log --oneline -5
```
Recent messages are short and lowercase (e.g. "golangci-lint", "Update metadata.lua", "token"). Match that style — a one-liner like `bump to vX.Y.Z` is plenty.

If the user has other staged changes they want shipped in this release, fold them into the same commit. Otherwise, commit only `metadata.lua` and `Makefile`.

Always include the trailer:
```
Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
```

### 6. Push

```bash
git push
```

If the branch has no upstream, use `-u origin <branch>`. If push is rejected (someone else pushed in the meantime), pull and retry — do NOT force-push.

### 7. Create the GitHub release

```bash
gh release create vX.Y.Z --title vX.Y.Z --notes ""
```

**Do not pass binary paths.** CI builds darwin-amd64, darwin-arm64, and linux-amd64 from the tag and attaches them automatically. (We learned this the hard way — the user corrected: "github-ci会自己生成的".)

Before running, confirm the tag doesn't already exist:
```bash
git ls-remote origin refs/tags/vX.Y.Z
```
If it does, stop and ask the user — the release likely already happened, or there's a stale tag to clean up first (`git push origin :refs/tags/vX.Y.Z` and `gh release delete vX.Y.Z`).

`gh release create` creates the tag and the release in one step and prints the release URL.

### 8. Report

Print the release URL and a one-line summary. Done.

## Why these guardrails

- **`metadata.lua` is the version source of truth** because it's what `mise`/`vfox` see. Skipping the bump means the released plugin reports the wrong version to users.
- **No local binary uploads** because CI is the canonical builder — locally-built binaries can have different toolchain/timestamps and confuse downstream consumers.
- **No force-push, no auto-rebase** because the cost of clobbering someone else's work or the user's in-flight branch is much higher than the cost of pausing to ask.
- **Cross-checking releases vs. metadata.lua in step 2** catches the case where the version was bumped on origin but the local clone is stale — releasing a duplicate or backward version is annoying to undo.

## Quick command reference

```bash
# Sync
git fetch && git pull --ff-only && git status

# Inspect
grep '^  version' metadata.lua
git tag --sort=-v:refname | head -3
gh release list --limit 3

# Bump (after confirming version with user)
# … edit metadata.lua AND Makefile (VERSION=X.Y.Z) …
git add metadata.lua Makefile
git commit -m "bump to vX.Y.Z"   # plus Co-Authored-By trailer
git push

# Release
git ls-remote origin refs/tags/vX.Y.Z   # must be empty
gh release create vX.Y.Z --title vX.Y.Z --notes ""
```
