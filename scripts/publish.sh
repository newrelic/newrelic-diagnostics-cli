#!/bin/bash

set -e

echo "Running build script"
CGO_ENABLED=0 ./scripts/build.sh

VERSION_NUMBER=$BUILD_NUMBER

if [ -z "$BUILD_NUMBER" ]
  then
    echo "No arguments supplied"
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
#upload to artifactory:
if [ "$VERSION_NUMBER" == "$BUILD_NUMBER" ]
  then 
 
  BUILDENV=staging
# now copy file to shared mounted folder so it can be uploaded to artifactory
echo "Copying $ZIPFILENAME to shared volume"
cp $ZIPFILENAME sharedfolder/

fi

