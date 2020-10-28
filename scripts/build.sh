#!/bin/sh
set -e

EXENAME=nrdiag

mkdir -p bin/mac
mkdir -p bin/win
mkdir -p bin/linux

VERSION_NUMBER=$BUILD_NUMBER
CONFIG_PATH="github.com/newrelic/NrDiag/config"

if [ -z "$BUILD_NUMBER" ]; then
  echo "No arguments supplied for BUILD_NUMBER"
  VERSION_NUMBER="DEVELOP"
  # Development mode, send usage data to staging endpoint
  USAGE_ENDPOINT="${STAGING_USAGE_ENDPOINT}"
  # Development mode, send attachments to staging endpoint
  ATTACHMENT_ENDPOINT="${STAGING_ATTACHMENT_ENDPOINT}"
  HABERDASHER_URL="${STAGING_HABERDASHER_URL}"
else
  # Build number is present, send usage data to production endpoint
  USAGE_ENDPOINT="${PROD_USAGE_ENDPOINT}"
  ATTACHMENT_ENDPOINT="${PROD_ATTACHMENT_ENDPOINT}"
  HABERDASHER_URL="${PROD_HABERDASHER_URL}"
fi

VERSION=$(cat majorMinorVersion.txt | awk -F'=' '{print$2}')

BUILD_TIMESTAMP=$(date -u '+%Y-%m-%d_%I:%M:%S%p')
LDFLAGS="-X ${CONFIG_PATH}.Version=${VERSION}.${VERSION_NUMBER} -X ${CONFIG_PATH}.BuildTimestamp=${BUILD_TIMESTAMP} -X ${CONFIG_PATH}.UsageEndpoint=${USAGE_ENDPOINT} -X ${CONFIG_PATH}.AttachmentEndpoint=${ATTACHMENT_ENDPOINT} -X ${CONFIG_PATH}.HaberdasherURL=${HABERDASHER_URL}"

# Set version based on version.txt file and auto version number
echo "Build version is $VERSION.$VERSION_NUMBER"
echo "Buildstamp is $BUILD_TIMESTAMP"

echo "Running go get -t ./..."
$(go get -t ./...)

echo "running GOOS=windows go get -t ./..."
$(GOOS=windows go get -t ./...)

echo "Building Mac x64 $EXENAME"
GOOS=darwin GOARCH=amd64 go build -o "$EXENAME" -ldflags "$LDFLAGS"
$(mv "$EXENAME" "bin/mac/${EXENAME}_x64")

echo "Building Linux 386"
GOOS=linux GOARCH=386 go build -o "$EXENAME" -ldflags "$LDFLAGS"
$(mv "$EXENAME" "bin/linux/$EXENAME")

echo "Building Linux x64"
GOOS=linux GOARCH=amd64 go build -o "$EXENAME" -ldflags "$LDFLAGS"
$(mv "$EXENAME" "bin/linux/${EXENAME}_x64")

echo "Building Windows 386"
GOOS=windows GOARCH=386 go build -o "$EXENAME" -ldflags "$LDFLAGS"
$(mv "$EXENAME" "bin/win/$EXENAME.exe")

echo "Building Windows x64"
GOOS=windows GOARCH=amd64 go build -o "$EXENAME" -ldflags "$LDFLAGS"
$(mv "$EXENAME" "bin/win/${EXENAME}_x64.exe")

