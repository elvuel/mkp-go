param(
    [string]$Arch = $(if ($env:GOARCH) { $env:GOARCH } else { "amd64" })
)

$serverRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$mainPath = Join-Path $serverRoot "main.go"

$env:GOOS = "windows"
$env:GOARCH = $Arch

$output = Join-Path $serverRoot "mkp-server.exe"
go build -ldflags="-s -w" -o $output $mainPath
Write-Host "Built $output"
