#!/usr/bin/env bash

# Install the New Relic Diagnostics CLI.
# https://github.com/newrelic/newrelic-diagnostics-cli
#
# Dependencies: curl, cut, tar, gzip
#
# The binary location can be passed in via DESTDIR, otherwise it installs to pwd.
#

set -e

# ------------------------- INIT

UNAME=($(uname -ms))
PLAT=${UNAME[0]}
ARCH=${UNAME[1]}
if [ -z "${BASE_URL}" ]; then
    BASE_URL="https://download.newrelic.com"
fi
MANUAL_INSTALL_DOC="https://docs.newrelic.com/docs/new-relic-solutions/solve-common-issues/diagnostics-cli-nrdiag/run-diagnostics-cli-nrdiag/"
DESTDIR="${DESTDIR:-${PWD}}"

# ------------------------- FUNCTIONS

checkOS() {
    case ${PLAT} in
    Linux) ;;
    *)
        echo "This operating system is not supported by this installation script. Please install the New Relic Diagnostics CLI manually."
        echo "${MANUAL_INSTALL_DOC}"
        exit 1
        ;;
    esac

    case ${ARCH} in
    arm64 | aarch64)
        ARCH="arm64"
        ;;
    x86_64)
        ARCH="x64"
        ;;
    *)
        echo "This machine architecture is not supported. The supported architectures are x86_64 and arm64."
        exit 1
        ;;
    esac
}

checkReqs() {
    for x in cut tar gzip sudo; do
        which $x >/dev/null || (
            echo "Unable to continue.  Please install $x before proceeding."
            exit 1
        )
    done
}

checkAndInstallCurl() {
    local DISTRO=$(cat /etc/issue /etc/system-release /etc/redhat-release /etc/os-release 2>/dev/null | grep -m 1 -Eo "(Ubuntu|Amazon|CentOS|Debian|Red Hat|SUSE)" || true)

    local IS_CURL_INSTALLED=$(which curl | wc -l)
    if [ ${IS_CURL_INSTALLED} -eq 0 ]; then
        echo "curl is required to install, please confirm Y/N to install (default Y): "
        read -r CONFIRM_CURL
        if [ "${CONFIRM_CURL}" == "Y" ] || [ "${CONFIRM_CURL}" == "y" ] || [ "${CONFIRM_CURL}" == "" ]; then
            if [ "${DISTRO}" == "Ubuntu" ] || [ "${DISTRO}" == "Debian" ]; then
                sudo apt-get update
                sudo apt-get install curl -y
            elif [ "${DISTRO}" == "Amazon" ] || [ "${DISTRO}" == "CentOS" ] || [ "${DISTRO}" == "Red Hat" ]; then
                sudo yum install curl -y
            elif [ "${DISTRO}" == "SUSE" ]; then
                sudo zypper -n install curl
            else
                echo "Unable to continue. Please install curl manually before proceeding."
                exit 1
            fi
        else
            echo "Unable to continue without curl. Please install curl before proceeding."
            exit 1
        fi
    fi
}

getLatestVersion() {
    VERSION=$(curl -sL ${BASE_URL}/nrdiag/version.txt)
}

checkConnectivity() {
    curl --connect-timeout 10 -IsL "${BASE_URL}" >/dev/null || (echo "Cannot connect to ${BASE_URL} to download the New Relic Diagnostics CLI. Check your firewall settings. If you are using a proxy, make sure you have set the HTTPS_PROXY environment variable." && exit 1)
}

setupDirectories() {
    if [ ! -d "${DESTDIR}" ]; then
        mkdir -m 755 -p "${DESTDIR}"
    fi
    SCRATCH=$(mktemp -d || mktemp -d -t 'tmp')
    cd "${SCRATCH}"

}

downloadAndExtract() {
    echo "Installing New Relic Diagnostics CLI v${VERSION}"
    curl -sL --retry 3 "${BASE_URL}/nrdiag/nrdiag_${VERSION}_${PLAT}_${ARCH}.tar.gz" | tar -xz
}

moveToDestDir() {
    if [ "$UID" != "0" ]; then
        echo "Installing to ${DESTDIR} using sudo"
        mv "nrdiag_${ARCH}" "${DESTDIR}/nrdiag"
        chmod +x "${DESTDIR}/nrdiag"
        sudo chown root:0 "${DESTDIR}/nrdiag"
    else
        echo "Installing to ${DESTDIR}"
        mv "nrdiag_${ARCH}" "${DESTDIR}/nrdiag"
        chmod +x "${DESTDIR}/nrdiag"
        chown root:0 "${DESTDIR}/nrdiag"
    fi
}

cleanup() {
    rm -r "${SCRATCH}"
}

error() {
    echo "An error occurred installing the tool."
    echo "The contents of the directory ${SCRATCH} have been left in place to help to debug the issue."
}

# ------------------------- MAIN

echo "Starting installation."
trap error ERR
checkOS
checkReqs
checkAndInstallCurl
checkConnectivity
getLatestVersion
setupDirectories
downloadAndExtract
moveToDestDir
cleanup
exit 0
