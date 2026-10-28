param([string]$BaseUrl = 'http://127.0.0.1:8080')
$ErrorActionPreference = 'Stop'
Push-Location (Split-Path $PSScriptRoot -Parent)
try { & go run ./cmd/dev smoke -url $BaseUrl; if ($LASTEXITCODE -ne 0) { throw 'Smoke failed' } }
finally { Pop-Location }
