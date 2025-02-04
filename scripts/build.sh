#!/bin/sh
set -e

TEMPNAME=newrelic-diagnostics-cli
EXENAME=nrdiag

mkdir -p bin/mac
mkdir -p bin/win
mkdir -p bin/linux

VERSION_NUMBER=$BUILD_NUMBER
CONFIG_PATH="github.com/newrelic/newrelic-diagnostics-cli/config"

if [ -z "$BUILD_NUMBER" ]; then
  echo "No arguments supplied for BUILD_NUMBER"
  VERSION_NUMBER="DEVELOP"
  # Development mode, send usage data to staging endpoint
  US_USAGE_ENDPOINT="${STAGING_USAGE_ENDPOINT}"
  # Development mode, send attachments to staging endpoint
  US_ATTACHMENT_ENDPOINT="${STAGING_ATTACHMENT_ENDPOINT}"
  US_HABERDASHER_URL="${STAGING_HABERDASHER_URL}"
  EU_USAGE_ENDPOINT="${STAGING_USAGE_ENDPOINT}"
  EU_ATTACHMENT_ENDPOINT="${STAGING_ATTACHMENT_ENDPOINT}"
  EU_HABERDASHER_URL="${STAGING_HABERDASHER_URL}"
else
  # Build number is present, send usage data to production endpoint
  echo "send usage data to Haberdasher production"
  US_USAGE_ENDPOINT="${PROD_USAGE_ENDPOINT}"
  US_ATTACHMENT_ENDPOINT="${PROD_ATTACHMENT_ENDPOINT}"
  US_HABERDASHER_URL="${PROD_HABERDASHER_URL}"
  EU_USAGE_ENDPOINT="${EU_USAGE_ENDPOINT}"
  EU_ATTACHMENT_ENDPOINT="${EU_ATTACHMENT_ENDPOINT}"
  EU_HABERDASHER_URL="${EU_HABERDASHER_URL}"
fi

VERSION=$(cat releaseVersion.txt | awk -F'majorMinor=' '{printf$2}')

BUILD_TIMESTAMP=$(date -u '+%Y-%m-%d_%I:%M:%S%p')
LDFLAGS="-s -w -X ${CONFIG_PATH}.Version=${VERSION}.${VERSION_NUMBER} -X ${CONFIG_PATH}.BuildTimestamp=${BUILD_TIMESTAMP} -X ${CONFIG_PATH}.USUsageEndpoint=${US_USAGE_ENDPOINT} -X ${CONFIG_PATH}.USAttachmentEndpoint=${US_ATTACHMENT_ENDPOINT} -X ${CONFIG_PATH}.USHaberdasherURL=${US_HABERDASHER_URL} -X ${CONFIG_PATH}.EUUsageEndpoint=${EU_USAGE_ENDPOINT} -X ${CONFIG_PATH}.EUAttachmentEndpoint=${EU_ATTACHMENT_ENDPOINT} -X ${CONFIG_PATH}.EUHaberdasherURL=${EU_HABERDASHER_URL}"

# Set version based on version.txt file and auto version number
echo "Build version is $VERSION.$VERSION_NUMBER"
echo "Buildstamp is $BUILD_TIMESTAMP"

echo "Running go get -t ./..."
$(go get -t ./...)

echo "running GOOS=windows go get -t ./..."
$(GOOS=windows go get -t ./...)

echo "Building Mac x64 $EXENAME"
GOOS=darwin GOARCH=amd64 go build -o "$TEMPNAME" -ldflags "$LDFLAGS"
$(mv "$TEMPNAME" "bin/mac/${EXENAME}_x64")

echo "Building Mac arm64 $EXENAME"
GOOS=darwin GOARCH=arm64 go build -o "$TEMPNAME" -ldflags "$LDFLAGS"
$(mv "$TEMPNAME" "bin/mac/${EXENAME}_arm64")

echo "Building Linux x64"
GOOS=linux GOARCH=amd64 go build -o "$TEMPNAME" -ldflags "$LDFLAGS"
$(mv "$TEMPNAME" "bin/linux/${EXENAME}_x64")

echo "Building Linux arm64"
GOOS=linux GOARCH=arm64 go build -o "$TEMPNAME" -ldflags "$LDFLAGS"
$(mv "$TEMPNAME" "bin/linux/${EXENAME}_arm64")

echo "Building Windows 386"
GOOS=windows GOARCH=386 go build -o "$TEMPNAME" -ldflags "$LDFLAGS"
$(mv "$TEMPNAME" "bin/win/$EXENAME.exe")

echo "Building Windows x64"
GOOS=windows GOARCH=amd64 go build -o "$TEMPNAME" -ldflags "$LDFLAGS"
$(mv "$TEMPNAME" "bin/win/${EXENAME}_x64.exe")

echo "Building Windows arm64"
GOOS=windows GOARCH=arm64 go build -o "$TEMPNAME" -ldflags "$LDFLAGS"
$(mv "$TEMPNAME" "bin/win/${EXENAME}_arm64.exe")
