#!/bin/bash
set -e

CURRENT=$(cat releaseVersion.txt | awk -F'currentReleaseVersion=' '{printf$2}')
CURRENTZIP="nrdiag_${CURRENT}.zip"
echo "my currentzip:"
echo nrdiag_${CURRENT}.zip
echo $CURRENTZIP
PREV=$(cat releaseVersion.txt | awk -F'prevReleaseVersion=' '{printf$2}')
PREVZIP="nrdiag_${PREV}.zip"
echo "my prevzip:"
echo nrdiag_${PREV}.zip
aws s3 rm s3://${{ secrets.S3_BUCKET }}/nrdiag/test/${CURRENTZIP}
echo "current release was removed"
aws s3 cp s3://${{ secrets.S3_BUCKET }}/nrdiag/test/${PREVZIP} s3://${{ secrets.S3_BUCKET }}/nrdiag/test/nrdiag_latest.zip
echo "previous release is now latest zip"
echo ${PREV} >> version.txt
aws s3 cp version.txt s3://${{ secrets.S3_BUCKET }}/nrdiag/test/version.txt