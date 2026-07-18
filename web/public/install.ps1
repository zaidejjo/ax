# ax — Smart Install Script (PowerShell)
# https://github.com/zaidejjo/ax
# Detects architecture and installs the latest Windows binary.
#Requires -Version 5.1

$repo = "zaidejjo/ax"
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$releases = "https://api.github.com/repos/$repo/releases/latest"

Write-Host "==> Fetching latest release..." -ForegroundColor Cyan
$tag = (Invoke-RestMethod -Uri $releases).tag_name

$fileName = "ax_windows_$arch.zip"
$downloadUrl = "https://github.com/$repo/releases/download/$tag/$fileName"
$zipPath = Join-Path $env:TEMP "ax.zip"
$extractPath = Join-Path $env:TEMP "ax"

Write-Host "==> Downloading $fileName..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath

Write-Host "==> Extracting..." -ForegroundColor Cyan
Expand-Archive -Path $zipPath -DestinationPath $extractPath -Force

# Install to a directory in PATH
$installDir = "$env:LOCALAPPDATA\ax"
if (!(Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir -Force }

Move-Item -Path "$extractPath\ax.exe" -Destination "$installDir\ax.exe" -Force

# Add to User PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
  $env:Path = "$env:Path;$installDir"
}

Write-Host "==> ax installed to $installDir" -ForegroundColor Green
Write-Host "==> Restart your terminal or run: ax" -ForegroundColor Cyan
