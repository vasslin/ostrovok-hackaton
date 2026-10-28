param([ValidateSet('demo','live')][string]$Mode = 'demo')
$ErrorActionPreference = 'Stop'
Push-Location (Split-Path $PSScriptRoot -Parent)
try { & go run ./cmd/dev run -mode $Mode; if ($LASTEXITCODE -ne 0) { throw 'Local stack failed' } }
finally { Pop-Location }
