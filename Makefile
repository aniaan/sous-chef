BINARY_NAME=sous-chef
VERSION=0.0.2
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: build clean release

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -rf lua/bin/

release:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 main.go
