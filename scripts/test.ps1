param([switch]$Race, [switch]$Cover)
$ErrorActionPreference = 'Stop'
Push-Location (Split-Path $PSScriptRoot -Parent)
try {
    $testFlags = @()
    if ($Race) { $testFlags += '-race' }
    if ($Cover) { $testFlags += '-cover' }
    & go run ./cmd/dev test @testFlags
    if ($LASTEXITCODE -ne 0) { throw 'Verification failed' }
} finally { Pop-Location }
