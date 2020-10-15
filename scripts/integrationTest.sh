#!/bin/bash
set -e
echo "Running go get -t"
go get -t


# Individual tests can be run by name by running   ./integrationTest.sh 'JavaVersionPresent,RunningPHPDaemon'

EXENAME=nrdiag
BUILD_TIMESTAMP=`date -u '+%Y-%m-%d_%I:%M:%S%p'`
# Send usage data to staging endpoint
USAGE_ENDPOINT="${STAGING_USAGE_ENDPOINT}"
ATTACHMENT_ENDPOINT="${STAGING_ATTACHMENT_ENDPOINT}"
CONFIG_PATH="github.com/newrelic/NrDiag/config"

echo "Building linux x86"
GOOS=linux GOARCH=386 go build -o "$EXENAME" -ldflags "-X ${CONFIG_PATH}.Version=INTEGRATION -X ${CONFIG_PATH}.BuildTimestamp=${BUILD_TIMESTAMP} -X ${CONFIG_PATH}.UsageEndpoint=${USAGE_ENDPOINT} -X ${CONFIG_PATH}.AttachmentEndpoint=${ATTACHMENT_ENDPOINT}"
mkdir -p integration/bin/linux
$(mv "$EXENAME" "integration/bin/linux/$EXENAME")
echo "Running integration tests"
go test integration/integration_test.go integration/integration_test_timer.go integration/dockerHelper_test.go -timeout 2h -v --tags=integration -testNames=$1
