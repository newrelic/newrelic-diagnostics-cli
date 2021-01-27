#!/bin/bash
set -e

CURRENT=$(cat releaseVersion.txt | awk -F'currentReleaseVersion=' '{printf$2}')
CURRENTZIP="nrdiag_${CURRENT}.zip"

PREV=$(cat releaseVersion.txt | awk -F'prevReleaseVersion=' '{printf$2}')
PREVZIP="nrdiag_${PREV}.zip"

echo "aws version is:"
aws --version

echo "deleting current release zip file ..."
aws s3 rm s3://${S3_BUCKET}/nrdiag/test/${CURRENTZIP}

echo "replacing nrdiag_latest.zip to be the previous one ..."
#--copy-props metadata-directive parameter is to avoid calling GetObjectTagging permission
aws s3 cp --copy-props metadata-directive s3://${S3_BUCKET}/nrdiag/test/${PREVZIP} s3://${S3_BUCKET}/nrdiag/test/nrdiag_latest.zip

echo "deleting version.txt ..."
aws s3 rm s3://${S3_BUCKET}/nrdiag/test/version.txt
echo "creating a new version.txt inserting the previous version ..."
echo ${PREV} >> version.txt
aws s3 cp version.txt s3://${S3_BUCKET}/nrdiag/test/version.txt
echo "version is now:"
aws s3 cp s3://nr-downloads-main/nrdiag/test/version.txt -