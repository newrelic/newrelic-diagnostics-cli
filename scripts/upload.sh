#!/usr/bin/env bash

# This script runs with the following env values passed in and mounts the current working directory into the container:
#    $ docker run --rm -e S3_BUCKET -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e BUILD_NUMBER \
#      -v $PWD/production:/root/go/src/github.com/newrelic/newrelic-diagnostics-cli/sharedfolder nrdiag-build ./scripts/upload.sh
# To check files from command line:
#    $ AWS_ACCESS_KEY_ID=abc AWS_SECRET_ACCESS_KEY=123 aws s3 ls s3://${S3_BUCKET}/nrdiag/

set -e

# ------------------------- INIT
BUILD_NUMBER_PATTERN="^[0-9]+$"
VERSION=$(cat releaseVersion.txt | awk -F'currentReleaseVersion=' '{printf$2}')
PREV_VERSION=$(cat releaseVersion.txt | awk -F'prevReleaseVersion=' '{printf$2}')
ZIPFILENAME="nrdiag_${VERSION}.zip"
SCRATCH_DIR="nrdiag_scratch_dir"
BASE_DIR="nrdiag"

# ------------------------- FUNCTIONS

checkBuildNumber() {
  if [[ ! ${BUILD_NUMBER} =~ ${BUILD_NUMBER_PATTERN} ]]; then
    echo "The input passed for github.event.release.name was not a valid number. The release process did not run."
    exit 1
  fi
}

build() {
  echo "Running build script"
  sh ./scripts/build.sh
  if [[ $? -ne 0 ]]; then
    echo "Build failed, aborting release"
    exit 1
  fi
  cp -r ./bin/* ${SCRATCH_DIR}/${BASE_DIR}
}

prepDirs() {
  mkdir -p ${SCRATCH_DIR}/${BASE_DIR}
}

createVersionText() {
  echo "${VERSION}" >>${SCRATCH_DIR}/${BASE_DIR}/version.txt
}

copyLicenseToZip() {
  cp ./licenses/* ${SCRATCH_DIR}/${BASE_DIR}
}

createZip() {
  echo "Creating zipfile ${ZIPFILENAME}"
  cd ./${SCRATCH_DIR}
  zip -r ${ZIPFILENAME} ${BASE_DIR}
  mv ${ZIPFILENAME} ./${BASE_DIR}
  cd ..
}

cleanUpBeforeUpload() {
  rm ${SCRATCH_DIR}/${BASE_DIR}/LICENSE*
  rm ${SCRATCH_DIR}/${BASE_DIR}/README.txt
  rm -rf ${SCRATCH_DIR}/${BASE_DIR}/linux
  rm -rf ${SCRATCH_DIR}/${BASE_DIR}/mac
  rm -rf ${SCRATCH_DIR}/${BASE_DIR}/win
}

createPlatformSpecificArchive() {
  local plat=$1
  local arch=$2
  echo "Creating platform specific archive for $plat $arch"

  case ${plat} in
  windows)
    cd ./bin/win
    if [[ "${arch}" == "x86" ]]; then
      zip ../../${SCRATCH_DIR}/${BASE_DIR}/nrdiag_${VERSION}_Windows_${arch}.zip ./nrdiag.exe
    else
      zip ../../${SCRATCH_DIR}/${BASE_DIR}/nrdiag_${VERSION}_Windows_${arch}.zip ./nrdiag_${arch}.exe
    fi
    cd ../..
    ;;
  linux)
    cd ./bin/linux
    tar czvf ../../${SCRATCH_DIR}/${BASE_DIR}/nrdiag_${VERSION}_Linux_${arch}.tar.gz ./nrdiag_${arch}
    cd ../..
    ;;

  *)
    echo "Unknown platform: ${plat}"
    echo "Aborting release"
    exit 1
    ;;
  esac
}

uploadToAws() {
  echo "Uploading to download.newrelic.com"
  cd ./${SCRATCH_DIR}/${BASE_DIR}
  for file in *; do
    aws s3 cp ${file} s3://${S3_BUCKET}/${BASE_DIR}/${file}
  done
  aws s3 cp s3://${S3_BUCKET}/${BASE_DIR}/nrdiag_${VERSION}.zip s3://${S3_BUCKET}/${BASE_DIR}/nrdiag_latest.zip --copy-props none
  cd ../..
}

removeOldestFromAws() {
  echo "Finding all files for oldest release on download.newrelic.com"
  local aws_sort_query='sort_by(Contents[?Key && (contains(Key, `.zip`) == `true` || contains(Key, `.tar.gz`) == `true`) && contains(Key, `nrdiag`) == `true` && contains(Key, `latest`) == `false` && contains(Key, `'"${VERSION}"'`) == `false` && contains(Key, `'"${PREV_VERSION}"'`) == `false`], &LastModified)[].[Key]'
  IFS=$'\n' read -r -d '' -a old_releases < <(aws s3api list-objects-v2 --bucket "${S3_BUCKET}" --prefix "${BASE_DIR}"/ --query "${aws_sort_query}" --output text && printf '\0')
  for rel in "${old_releases[@]}"; do
    aws s3 rm s3://${S3_BUCKET}/${rel}
  done
}

# ------------------------- MAIN

checkBuildNumber
prepDirs
build
createVersionText
copyLicenseToZip
createZip
createPlatformSpecificArchive linux x64
createPlatformSpecificArchive linux arm64
createPlatformSpecificArchive windows x86
createPlatformSpecificArchive windows x64
createPlatformSpecificArchive windows arm64
cleanUpBeforeUpload
uploadToAws
removeOldestFromAws

exit 0
