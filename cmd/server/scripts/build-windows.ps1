param(
    [string]$Arch = $(if ($env:GOARCH) { $env:GOARCH } else { "amd64" })
)

$serverRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$mainPath = Join-Path $serverRoot "."
$gitDate = git show -s --date=format:%Y%m%d --format=%cd HEAD
$gitHash = git rev-parse --short=8 HEAD
$version = "$gitDate-$gitHash"
Write-Host "Building version $version"
$ldflags = "-s -w -X main.Version=$version"

$env:GOOS = "windows"
$env:GOARCH = $Arch

$output = Join-Path $serverRoot "mkp-server.exe"
go build -ldflags="$ldflags" -o $output $mainPath
Write-Host "Built $output"
