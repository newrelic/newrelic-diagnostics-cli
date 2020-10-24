param([string]$tests)
# We can run a test by name by running powershell .\integrationTest_windows.ps1 'FailingCollector,SuccessfullCollector'

echo "Running go get -t"
go get -t


$EXENAME = "nrdiag"
$BUILD_TIMESTAMP = Get-Date -format s # 2017-09-06_03:28:19PM
echo "Building Windows x86"
$env:GOARCH='386'
go build -o $EXENAME -ldflags "-X github.com/newrelic/NrDiag/config.Version=INTEGRATION -X github.com/newrelic/NrDiag/config.BuildTimestamp=$BUILD_TIMESTAMP"
mkdir .\bin\win\ -ErrorAction Ignore
move-item -Force -Path .\$EXENAME -Destination .\bin\win\$EXENAME.exe 
echo "Building Windows x64"
$env:GOARCH='amd64'
go build -o $EXENAME -ldflags "-X github.com/newrelic/NrDiag/config.Version=INTEGRATION -X github.com/newrelic/NrDiag/config.BuildTimestamp=$BUILD_TIMESTAMP"
move-item -Force -Path .\nrdiag -Destination .\bin\win\"$EXENAME"_x64.exe

go test integration_test.go integration_test_timer.go dockerHelper_test.go -v -timeout 3h --tags=integration -testNames="$tests"

if ($LASTEXITCODE -eq 1) {
    exit $LASTEXITCODE
    }
                    
