$ErrorActionPreference = "Stop"

$repo = "Harsh-2002/Zeno"
$binary = "zeno"

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { Write-Error "32-bit not supported"; exit 1 }
$asset = "${binary}-windows-${arch}.exe"

Write-Host "-> Detected: windows/${arch}"

# Get latest release
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/${repo}/releases/latest"
$tag = $release.tag_name
Write-Host "-> Latest release: ${tag}"

# Download
$url = "https://github.com/${repo}/releases/download/${tag}/${asset}"
$installDir = "$env:LOCALAPPDATA\Zeno"
$dest = "${installDir}\${binary}.exe"

if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir | Out-Null }

Write-Host "-> Downloading ${asset}..."
Invoke-WebRequest -Uri $url -OutFile $dest

# Add to PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*${installDir}*") {
    [Environment]::SetEnvironmentVariable("Path", "${userPath};${installDir}", "User")
    Write-Host "-> Added ${installDir} to PATH"
}

Write-Host "-> Installed to ${dest}"
Write-Host "-> Done! Restart your terminal, then run: zeno"
