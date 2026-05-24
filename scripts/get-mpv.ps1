# get-mpv.ps1
# Downloads the portable mpv build and places it in assets/mpv/.
# Run once before building or testing footage.
#
# Usage (from repo root):
#   powershell -ExecutionPolicy Bypass -File scripts/get-mpv.ps1

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot  = Split-Path -Parent $scriptDir
$outDir    = Join-Path $repoRoot "assets\mpv"
$mpvExe    = Join-Path $outDir "mpv.exe"

if (Test-Path $mpvExe) {
    Write-Host "mpv already present at $mpvExe" -ForegroundColor Green
    exit 0
}

# mpv portable build for Windows x86_64
# Releases: https://sourceforge.net/projects/mpv-player-windows/files/
# This URL points to a stable release; update when a new version ships.
$mpvVersion = "2024-01-14"
$zipUrl = "https://sourceforge.net/projects/mpv-player-windows/files/release/mpv-x86_64-20240114-git-a39f37b.7z/download"

# Fallback: install via winget if 7-zip is not available
Write-Host "Looking for mpv via winget..." -ForegroundColor Cyan

$wingetMpv = Get-Command winget -ErrorAction SilentlyContinue
if ($wingetMpv) {
    Write-Host "Trying: winget install mpv.net" -ForegroundColor Cyan
    try {
        winget install --id mpv.net -e --accept-package-agreements --accept-source-agreements 2>&1
        $candidates = @(
            "$env:LOCALAPPDATA\Microsoft\WinGet\Packages\mpv.net*\mpv.exe",
            "C:\Program Files\mpv.net\mpv.exe",
            "C:\mpv\mpv.exe"
        )
        foreach ($c in $candidates) {
            $found = Get-ChildItem $c -ErrorAction SilentlyContinue | Select-Object -First 1
            if ($found) {
                Write-Host "Found mpv at $($found.FullName)" -ForegroundColor Green
                Write-Host "Creating assets/mpv/ symlink or copy..." -ForegroundColor Cyan
                New-Item -ItemType Directory -Force -Path $outDir | Out-Null
                Copy-Item $found.FullName $mpvExe -Force
                Write-Host "mpv copied to $mpvExe" -ForegroundColor Green
                exit 0
            }
        }
    } catch {
        Write-Host "winget install failed: $_" -ForegroundColor Yellow
    }
}

# Manual fallback: user copies mpv.exe themselves
Write-Host ""
Write-Host "Could not auto-install mpv." -ForegroundColor Yellow
Write-Host ""
Write-Host "Please install mpv manually and place mpv.exe in:" -ForegroundColor White
Write-Host "  $outDir" -ForegroundColor Cyan
Write-Host ""
Write-Host "Options:" -ForegroundColor White
Write-Host "  1. WinGet:  winget install mpv.net" -ForegroundColor Gray
Write-Host "     Then copy mpv.exe from the installed location to $outDir" -ForegroundColor Gray
Write-Host ""
Write-Host "  2. Scoop:   scoop install mpv" -ForegroundColor Gray
Write-Host "     Then copy mpv.exe from the scoop shims to $outDir" -ForegroundColor Gray
Write-Host ""
Write-Host "  3. Direct download: https://mpv.io/installation/" -ForegroundColor Gray
Write-Host "     Download the Windows build, extract, copy mpv.exe to $outDir" -ForegroundColor Gray
Write-Host ""
Write-Host "After placing mpv.exe, footage will find it automatically." -ForegroundColor White
exit 1
