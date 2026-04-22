#Requires -Version 5.1
<#
.SYNOPSIS
    Install the flow CLI on Windows.
.DESCRIPTION
    Downloads and installs the latest (or specified) version of flow to a local directory and adds it to PATH.
.PARAMETER Version
    Specific version to install (e.g. "v0.10.0"). Defaults to the latest release.
.PARAMETER InstallDir
    Directory to install the binary into. Defaults to "$env:LOCALAPPDATA\flow\bin".
.EXAMPLE
    irm https://raw.githubusercontent.com/flowexec/flow/main/scripts/install.ps1 | iex
.EXAMPLE
    .\install.ps1 -Version v0.10.0
#>
param(
    [string]$Version,
    [string]$InstallDir
)

$ErrorActionPreference = "Stop"

$Owner = "flowexec"
$Name = "flow"
$Binary = "flow"

function Get-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
    }
}

function Get-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Name/releases/latest"
    return $release.tag_name
}

$Arch = Get-Arch
if (-not $Version) {
    Write-Host "Fetching latest version..."
    $Version = Get-LatestVersion
}
if (-not $InstallDir) {
    $InstallDir = Join-Path $env:LOCALAPPDATA "flow\bin"
}

$DownloadUrl = "https://github.com/$Owner/$Name/releases/download/$Version/${Binary}_${Version}_windows_${Arch}.zip"
$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) "flow-install-$([System.Guid]::NewGuid().ToString('N'))"
$DownloadPath = Join-Path $TmpDir "$Binary.zip"

New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

Write-Host "Downloading $Binary $Version for windows/$Arch..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $DownloadPath -UseBasicParsing
} catch {
    Write-Error "Failed to download $DownloadUrl"
    exit 1
}

Write-Host "Installing $Binary $Version to $InstallDir..."
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Expand-Archive -Path $DownloadPath -DestinationPath $TmpDir -Force
Move-Item -Path (Join-Path $TmpDir "$Binary.exe") -Destination (Join-Path $InstallDir "$Binary.exe") -Force

# Clean up
Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue

# Add to user PATH if not already present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$InstallDir;$UserPath", "User")
    $env:Path = "$InstallDir;$env:Path"
    Write-Host "Added $InstallDir to user PATH."
}

Write-Host "$Binary was installed successfully to $InstallDir\$Binary.exe"
Write-Host "Run '$Binary --help' to get started."
