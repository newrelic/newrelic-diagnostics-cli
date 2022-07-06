$EXENAME="newrelic-diagnostics-cli"
echo "Running go get -t ./..."
go get -t ./...
go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@latest
go get github.com/onsi/ginkgo/v2

go build 
echo "Running unit tests"
#this runs all test in any directory that has a *_test.go file
ginkgo -r --keep-going --no-color
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

# This can be run from the madhatter-build dockerfile with
# docker run --rm madhatter-build:latest powershell .\scripts\test_windows.ps1