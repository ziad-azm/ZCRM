#Requires -Version 5.1
<#
.SYNOPSIS
  Runs golang-migrate against the local ZCRM database.
.EXAMPLE
  ./scripts/migrate.ps1 up
  ./scripts/migrate.ps1 down 1
  ./scripts/migrate.ps1 version
#>
param(
    [Parameter(Mandatory = $true, ValueFromRemainingArguments = $true)]
    [string[]]$MigrateArgs
)

$ErrorActionPreference = 'Stop'

if ($env:DATABASE_URL) { $dsn = $env:DATABASE_URL }
else { $dsn = 'postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable' }

$migrationsDir = Join-Path $PSScriptRoot '..\migrations'

& migrate -path $migrationsDir -database $dsn @MigrateArgs
exit $LASTEXITCODE
