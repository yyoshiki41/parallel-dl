PKGS=$(shell go list ./... | grep -v examples)

.PHONY: all help get-deps test test-ci

all: help

help:
	@echo "make get-deps      #=> Install dependencies"
	@echo "make test          #=> Run tests"

get-deps:
	@echo "go get leaktest"
	@go get github.com/fortytw2/leaktest

test:
	go test $(PKGS)

test-ci:
	@echo "go test"
	@go test -race -coverprofile=coverage.txt -covermode=atomic $(PKGS)
