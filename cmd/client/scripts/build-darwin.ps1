param(
    [string]$Arch = $(if ($env:GOARCH) { $env:GOARCH } else { "amd64" })
)

$clientRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$skillScripts = Join-Path $clientRoot "openclaw_skills\\mkp\\scripts"
$mainPath = Join-Path $clientRoot "."
$gitDate = git show -s --date=format:%Y%m%d --format=%cd HEAD
$gitHash = git rev-parse --short=8 HEAD
$version = "$gitDate-$gitHash"
$ldflags = "-s -w -X main.Version=$version"

$env:GOOS = "darwin"
$env:GOARCH = $Arch

New-Item -ItemType Directory -Force -Path $skillScripts | Out-Null
$output = Join-Path $skillScripts "mkp"

go build -ldflags="$ldflags" -o $output $mainPath
Write-Host "Built $output"
