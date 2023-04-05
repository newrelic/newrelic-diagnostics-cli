#!/usr/bin/env bash
#
# Rollback to the previous version of the New Relic Diagnostics CLI.
#

set -e

# ------------------------- INIT
VERSION_TO_REMOVE=$(cat releaseVersion.txt | awk -F'currentReleaseVersion=' '{printf$2}')
VERSION_TO_KEEP=$(cat releaseVersion.txt | awk -F'prevReleaseVersion=' '{printf$2}')
BASE_DIR="nrdiag"

# ------------------------- FUNCTIONS
checkS3BucketIsSet() {
    if [ -z "${S3_BUCKET}" ]; then
        echo "The S3 bucket variable was not passed in, aborting."
        exit 1
    fi
}

findAndRemove() {
    echo "Finding and removing all files for version ${VERSION_TO_REMOVE} on download.newrelic.com"
    local aws_sort_query='sort_by(Contents[?Key && (contains(Key, `.zip`) == `true` || contains(Key, `.tar.gz`) == `true`) && contains(Key, `nrdiag`) == `true` && contains(Key, `latest`) == `false` && contains(Key, `'"${VERSION_TO_REMOVE}"'`) == `true`], &LastModified)[].[Key]'
    IFS=$'\n' read -r -d '' -a files_to_remove < <(aws s3api list-objects-v2 --bucket "${S3_BUCKET}" --prefix "${BASE_DIR}"/ --query "${aws_sort_query}" --output text && printf '\0')
    for rel in "${files_to_remove[@]}"; do
        aws s3 rm s3://${S3_BUCKET}/${rel}
    done
}

fixVersionTxt() {
    echo "Deleting version.txt"
    aws s3 rm s3://${S3_BUCKET}/${BASE_DIR}/version.txt
    echo "Creating a new version.txt inserting the previous version"
    echo ${VERSION_TO_KEEP} >>version.txt
    aws s3 cp version.txt s3://${S3_BUCKET}/${BASE_DIR}/version.txt
    echo "Cleaning up local version.txt"
    rm ./version.txt
}

recreateLatestZip() {
    echo "Recreating nrdiag_latest.zip with version ${VERSION_TO_KEEP}"
    aws s3 cp s3://${S3_BUCKET}/${BASE_DIR}/nrdiag_${VERSION_TO_KEEP}.zip s3://${S3_BUCKET}/${BASE_DIR}/nrdiag_latest.zip --copy-props none
}

# ------------------------- MAIN
checkS3BucketIsSet
findAndRemove
fixVersionTxt
recreateLatestZip
