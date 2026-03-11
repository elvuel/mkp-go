param(
    [string]$Arch = $(if ($env:GOARCH) { $env:GOARCH } else { "amd64" })
)

$clientRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$skillScripts = Join-Path $clientRoot "openclaw_skills\\mkp\\scripts"
$mainPath = Join-Path $clientRoot "main.go"

$env:GOOS = "linux"
$env:GOARCH = $Arch

New-Item -ItemType Directory -Force -Path $skillScripts | Out-Null
$output = Join-Path $skillScripts "mkp"

go build -ldflags="-s -w" -o $output $mainPath
Write-Host "Built $output"
