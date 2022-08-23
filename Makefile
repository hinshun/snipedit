.PHONY: all snipedit

export GOBIN		:= $(shell dirname $(shell which go))
export GO		:= ${GOBIN}/go
export GOFMT		:= ${GOBIN}/gofmt
export GOLINT		:= ${GOBIN}/golangci-lint
export GO111MODULE	:= on
export GOSUMDB		:= off
export GOOS		:= linux
export GOARCH		:= amd64

GOPATH ?= $(HOME)/go

VERSION := $(shell git describe --tags --always --dirty)

all: snipedit

snipedit:
	# VERSION: ${VERSION}
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 ${GO} build ./cmd/snipedit

lint:
	${GOFMT} -d -l .
	${GO} vet ./...
	${GOLINT} run

unit-test:
	${GO} test -race -cover -count=1 -covermode=atomic -v ./...
