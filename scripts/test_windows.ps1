$EXENAME="newrelic-diagnostics-cli"

go mod download
go mod tidy
go build

#this runs all test in any directory that has a *_test.go file
go run github.com/onsi/ginkgo/v2/ginkgo -r --keep-going --no-color
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

# This can be run from the madhatter-build dockerfile with
# docker run --rm madhatter-build:latest powershell .\scripts\test_windows.ps1
