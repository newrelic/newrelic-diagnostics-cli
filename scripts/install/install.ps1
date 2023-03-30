# Install the New Relic Diagnostics CLI.
# https://github.com/newrelic/newrelic-diagnostics-cli
#
# Dependency: Powershell 5
#

Write-Output "`nStarting installation"

# Check OS
if ([System.Environment]::OSVersion.Platform -ne "Win32NT") {
    Write-Output "`nThis script is designed for Windows. Please see our documentation for other options to install the Diagnostics CLI on this host.`n`nhttps://docs.newrelic.com/docs/new-relic-solutions/solve-common-issues/diagnostics-cli-nrdiag/diagnostics-cli-nrdiag/"
    throw
}

# Check proc architecture
$arch = $Env:PROCESSOR_ARCHITECTURE.ToLower()
switch ($arch) {
    "x86" {Break}
    "amd64" {$arch = "x64"; Break}
    "arm64" {Break}
    Default {
        "This machine architecture is not supported. The supported architectures are amd64, x86, and arm64"
        throw
    }
}

# Check admin privileges
$p = New-Object System.Security.Principal.WindowsPrincipal([System.Security.Principal.WindowsIdentity]::GetCurrent())
if (!$p.IsInRole([System.Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw 'This script requires admin privileges to run and the current Windows PowerShell session is not running as Administrator. Start Windows PowerShell by using the Run as Administrator option, and then try running the script again.'
}

# Setup WebClient
[Net.ServicePointManager]::SecurityProtocol = 'tls12, tls';
$WebClient = New-Object System.Net.WebClient
if ($env:HTTPS_PROXY) {
    $WebClient.Proxy = New-Object System.Net.WebProxy($env:HTTPS_PROXY, $true)
}
$base_url = "https://download.newrelic.com"

# Get latest version
$version = $null
try {
    $version = $WebClient.DownloadString("${base_url}/nrdiag/version.txt").Trim();
}
catch {
    Write-Output "`nCould not download the latest version of the New Relic Diagnostics CLI.`n`nCheck your firewall settings. If you are using a proxy, make sure that you are able to access https://download.newrelic.com and that you have set the HTTPS_PROXY environment variable with your full proxy URL.`n"
    throw
}

# Clean up any existing nrdiag binaries in pwd
if (Test-Path -Path ".\nrdiag.exe" -PathType Leaf) {
    Write-Output "Removing existing nrdiag.exe binary"
    Remove-Item ".\nrdiag.exe"
}
if (Test-Path -Path ".\nrdiag_${arch}.exe" -PathType Leaf) {
    Write-Output "Removing existing nrdiag_${arch}.exe binary"
    Remove-Item ".\nrdiag_${arch}.exe"
}

# Check for existing zip
if (-not(Test-Path -Path "$env:TEMP\nrdiag_${version}_Windows_${arch}.zip" -PathType Leaf)) {
    # Download Zip if it doesn't already exist in temp
    try {
        $WebClient.DownloadFile("${base_url}/nrdiag/nrdiag_${version}_Windows_${arch}.zip", "$env:TEMP\nrdiag_${version}_Windows_${arch}.zip");
    }
    catch {
        Write-Output "`nCould not download the latest version of the New Relic Diagnostics CLI.`n`nCheck your firewall settings. If you are using a proxy, make sure that you are able to access https://download.newrelic.com and that you have set the HTTPS_PROXY environment variable with your full proxy URL.`n"
        throw
    }
}

# Expand zip to pwd
try {
    Write-Output "Installing New Relic Diagnostics CLI v${version} to ${pwd}"
    Expand-Archive -LiteralPath "$env:TEMP\nrdiag_${version}_Windows_${arch}.zip" -DestinationPath $pwd
}
catch {
    Write-Output "`nCould not extract the New Relic Diagnostics CLI.`n`nYou may try manually completing the installation by extracting the zip file located at $env:TEMP\nrdiag_${version}_Windows_${arch}.zip to $pwd`n"
    throw
}

# Rename to nrdiag
try {
    if ($arch -eq "x64" -or $arch -eq "arm64") {
        Rename-Item -Path "nrdiag_${arch}.exe" -NewName "nrdiag.exe"
    }
}
catch {
    Write-Output "`nCould not rename the New Relic Diagnostics CLI binary.`n"
    throw
}