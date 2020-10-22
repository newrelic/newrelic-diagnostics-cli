$EXENAME="NrDiag"
echo "Running go get -t ./..."
go get -t ./...
go get github.com/onsi/ginkgo/ginkgo # needed to pull down the ginkgo binary

go build 
echo "Running unit tests"
#this runs all test in any directory that has a *_test.go file
ginkgo -r -keepGoing -noColor
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

# This can be run from the madhatter-build dockerfile with
# docker run --rm madhatter-build:latest powershell .\scripts\test_windows.ps1