#!/bin/sh

set -e

EXENAME=newrelic-diagnostics-cli

# This looks for an import of "path" vs "path/filepath", see overview at top of https://golang.org/pkg/path/
if grep -R --include="*.go" -n "\"path\"" ./ *.go; then
    echo 'found suspected import of "path" instead of "path/filepath", see overview at top of https://golang.org/pkg/path/'
    exit 1
fi

go get -t ./...
go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@latest
go get github.com/onsi/ginkgo/v2
go build

ginkgo --skip-package dotnet/agent,dotnet/requirements,dotnet/env --no-color --keep-going -r -timeout=1h

# This can be run from the madhatter-build dockerfile with
# docker run --rm madhatter-build:latest ./publish.sh
