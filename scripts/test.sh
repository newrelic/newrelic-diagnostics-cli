#!/bin/sh

set -e

EXENAME=nr-diagnostics

# This looks for an import of "path" vs "path/filepath", see overview at top of https://golang.org/pkg/path/
if grep -R --include="*.go" -n "\"path\"" ./ *.go; then
    echo 'found suspected import of "path" instead of "path/filepath", see overview at top of https://golang.org/pkg/path/'
    exit 1
fi

go get -t ./...
go get github.com/onsi/ginkgo/ginkgo
go build

ginkgo -skipPackage dotnet/agent,dotnet/requirements,dotnet/env -noColor -keepGoing -r 

# This can be run from the madhatter-build dockerfile with
# docker run --rm madhatter-build:latest ./publish.sh
