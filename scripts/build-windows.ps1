$ErrorActionPreference = 'Stop'

$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'

New-Item -ItemType Directory -Force -Path bin | Out-Null
go build -trimpath -ldflags='-s -w' -o bin/gosysmon.exe ./cmd/gosysmon
