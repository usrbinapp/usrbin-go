SHELL := /bin/bash -o pipefail
VERSION ?=`git describe --tags`
DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
VERSION_PACKAGE = github.com/usrbinapp/usrbin-go/pkg/version

CURRENT_USER := $(shell id -u -n)
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

.PHONY: all
all: build

.PHONY: build
build:
	go build .

.PHONY: fmt
fmt:
	go fmt ./pkg/... .

.PHONY: vet
vet:
	go vet ./pkg/...

.PHONY: test
test: fmt vet
	go test ./pkg/...

