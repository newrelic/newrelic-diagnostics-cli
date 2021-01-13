#!/bin/bash

# This script runs with the following env values passed in and mounts the current working directory into the container:
# docker run --rm -e S3_BUCKET  -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e BUILD_NUMBER -v $PWD/production:/root/go/src/github.com/newrelic/newrelic-diagnostics-cli/sharedfolder nrdiag-build ./scripts/upload.sh

set -e

echo "Running build script"
./scripts/build.sh

VERSION_NUMBER=$BUILD_NUMBER

if [ -z "$BUILD_NUMBER" ]
  then
    echo "No arguments supplied for BUILD_NUMBER"
    VERSION_NUMBER="DEVELOP"
fi

VERSION=`cat releaseVersion.txt| awk -F'majorMinor=' '{printf$2}'`

ZIPFILENAME="nrdiag_${VERSION}.${BUILD_NUMBER}.zip"

echo "Creating zipfile $ZIPFILENAME"
cd bin
cp ../licenses/* ./
echo "${VERSION}.${BUILD_NUMBER}" >> version.txt
mkdir ../nrdiag/
cp -r ./* ../nrdiag/
cd ../
zip -r $ZIPFILENAME nrdiag/

if [ "$VERSION_NUMBER" == "$BUILD_NUMBER" ]
  then 
# echo "Uploading to artifactory"
#   BUILDENV=production
# echo "Copying $ZIPFILENAME to shared volume"
# cp $ZIPFILENAME sharedfolder/

echo "Uploading to Download.Newrelic.com"
ln -s $ZIPFILENAME nrdiag_latest.zip

aws s3 cp ${ZIPFILENAME} s3://${S3_BUCKET}/nrdiag/
aws s3 cp nrdiag_latest.zip s3://${S3_BUCKET}/nrdiag/
aws s3 cp bin/version.txt s3://${S3_BUCKET}/nrdiag/

fi

#AWS_ACCESS_KEY_ID=abc AWS_SECRET_ACCESS_KEY=123 aws s3 ls s3://bucketname/nrdiag/
