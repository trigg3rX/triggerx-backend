#!/bin/bash

# Go packages to install
go install go.uber.org/mock/mockgen@latest
go install github.com/axw/gocov/gocov@v1.1.0
go install github.com/matm/gocov-html/cmd/gocov-html@latest

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.7.1
