#!/bin/bash
set -e

CURRENT=$(cat releaseVersion.txt | awk -F'currentReleaseVersion=' '{printf$2}')
CURRENTZIP="nrdiag_${CURRENT}.zip"

PREV=$(cat releaseVersion.txt | awk -F'prevReleaseVersion=' '{printf$2}')
PREVZIP="nrdiag_${PREV}.zip"

echo "current aws version is:"
aws --version

echo "deleting current release zip file ..."
aws s3 rm s3://${S3_BUCKET}/nrdiag/${CURRENTZIP}

echo "deleting version.txt ..."
aws s3 rm s3://${S3_BUCKET}/nrdiag/version.txt
echo "creating a new version.txt inserting the previous version ..."
echo ${PREV} >> version.txt
aws s3 cp version.txt s3://${S3_BUCKET}/nrdiag/version.txt
echo "version is now:"
aws s3 cp s3://nr-downloads-main/nrdiag/version.txt -

echo "replacing nrdiag_latest.zip to be the previous one ..."
#once ubuntu-latest image upgrades their aws version to version 2, we'll need to update this command to use "--copy-props metadata-directive" arguments to avoid calling GetObjectTagging permission when copying a  file from bucket to bucket:
aws s3 cp s3://${S3_BUCKET}/nrdiag/${PREVZIP} s3://${S3_BUCKET}/nrdiag/nrdiag_latest.zip