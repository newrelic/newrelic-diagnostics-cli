#!/bin/sh

set -e

EXENAME=newrelic-diagnostics-cli

go get -t ./...
go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@latest
go get github.com/onsi/ginkgo/v2
go mod tidy
go build

ginkgo --skip-package dotnet/agent,dotnet/requirements,dotnet/env,dotnet/profiler --no-color --keep-going -r -timeout=1h

# This can be run from the madhatter-build dockerfile with
# docker run --rm madhatter-build:latest ./publish.sh
