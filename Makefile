PKGS=$(shell go list ./... | grep -v examples)

.PHONY: all help test test-ci

all: help

help:
	@echo "make test          #=> Run tests"

test:
	go test $(PKGS)

test-ci:
	@echo "go test"
	@go test -race -coverprofile=coverage.txt -covermode=atomic $(PKGS)
