# start.ps1 - interactive dev-ops menu for JojoOnePieceSimulator2.
#
# Wraps the same docker compose files apps/backend/Makefile uses
# (deployments/docker-compose.yml + docker-compose.dev.yml), plus goose
# migrations and a couple of local Go commands, behind one menu so you don't
# have to remember the exact compose/exec incantations during development.
#
# Usage: .\start.ps1            (interactive menu)
#        .\start.ps1 -Once up   (run one action non-interactively, then exit)

param(
    [string]$Once
)

$ErrorActionPreference = 'Stop'

$RepoRoot = $PSScriptRoot
$DeployDir = Join-Path $RepoRoot 'deployments'
$BackendDir = Join-Path $RepoRoot 'apps\backend'
$ComposeFiles = @(
    '-f', (Join-Path $DeployDir 'docker-compose.yml'),
    '-f', (Join-Path $DeployDir 'docker-compose.dev.yml')
)

function Invoke-Compose {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Args)
    & docker compose @ComposeFiles @Args
}

function Ensure-DockerRunning {
    try {
        docker info *> $null
        return $true
    } catch {
        Write-Host "Docker Desktop doesn't seem to be running." -ForegroundColor Yellow
        $start = Read-Host "Try to start it now? (y/N)"
        if ($start -eq 'y') {
            $dockerDesktop = "$env:ProgramFiles\Docker\Docker\Docker Desktop.exe"
            if (Test-Path $dockerDesktop) {
                Start-Process $dockerDesktop
                Write-Host "Launched Docker Desktop - waiting for it to come up..."
                $deadline = (Get-Date).AddSeconds(90)
                while ((Get-Date) -lt $deadline) {
                    Start-Sleep -Seconds 3
                    try { docker info *> $null; return $true } catch {}
                }
            }
            Write-Host "Docker still isn't responding. Start it manually and re-run this script." -ForegroundColor Red
            return $false
        }
        return $false
    }
}

function Ensure-PublicNet {
    $exists = docker network ls --format '{{.Name}}' | Select-String -SimpleMatch 'public-net'
    if (-not $exists) {
        Write-Host "Creating missing external network 'public-net'..."
        docker network create public-net | Out-Null
    }
}

function Action-UpAll {
    if (-not (Ensure-DockerRunning)) { return }
    Ensure-PublicNet
    Write-Host "Building and starting the full stack (postgres, redis, backend, frontend)..." -ForegroundColor Cyan
    Invoke-Compose up -d --build
    Invoke-Compose ps
}

function Action-UpService {
    param([string]$Service)
    if (-not (Ensure-DockerRunning)) { return }
    Ensure-PublicNet
    Write-Host "Building and starting '$Service'..." -ForegroundColor Cyan
    Invoke-Compose up -d --build $Service
}

function Action-Restart {
    param([string]$Service)
    if (-not (Ensure-DockerRunning)) { return }
    Write-Host "Restarting '$Service'..." -ForegroundColor Cyan
    Invoke-Compose restart $Service
}

function Action-Down {
    if (-not (Ensure-DockerRunning)) { return }
    Write-Host "Stopping the stack (containers only, volumes kept)..." -ForegroundColor Cyan
    Invoke-Compose down
}

function Action-Logs {
    param([string]$Service)
    if (-not (Ensure-DockerRunning)) { return }
    Write-Host "Tailing logs for '$Service' (Ctrl+C to stop)..." -ForegroundColor Cyan
    Invoke-Compose logs -f --tail=200 $Service
}

function Action-Ps {
    if (-not (Ensure-DockerRunning)) { return }
    Invoke-Compose ps
}

function Action-MigrateUp {
    if (-not (Ensure-DockerRunning)) { return }
    Write-Host "Applying pending migrations (goose up)..." -ForegroundColor Cyan
    Invoke-Compose exec backend sh -c 'goose -dir /migrations postgres "$DATABASE_URL" up'
}

function Action-MigrateDown {
    if (-not (Ensure-DockerRunning)) { return }
    Write-Host "Reverting the last migration (goose down)..." -ForegroundColor Cyan
    $confirm = Read-Host "This rolls back one migration - continue? (y/N)"
    if ($confirm -ne 'y') { return }
    Invoke-Compose exec backend sh -c 'goose -dir /migrations postgres "$DATABASE_URL" down'
}

function Action-MigrateStatus {
    if (-not (Ensure-DockerRunning)) { return }
    Invoke-Compose exec backend sh -c 'goose -dir /migrations postgres "$DATABASE_URL" status'
}

function Action-BackendTest {
    Write-Host "Running backend tests (go test ./...)..." -ForegroundColor Cyan
    Push-Location $BackendDir
    try { go test ./... } finally { Pop-Location }
}

function Action-Swagger {
    Write-Host "Regenerating swagger docs..." -ForegroundColor Cyan
    Push-Location $BackendDir
    try { go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/app/main.go -o docs --parseInternal } finally { Pop-Location }
}

function Action-Sqlc {
    Write-Host "Regenerating sqlc code..." -ForegroundColor Cyan
    Push-Location $BackendDir
    try { sqlc generate } finally { Pop-Location }
}

function Show-Menu {
    Write-Host ""
    Write-Host "=== JojoOnePieceSimulator2 - dev menu ===" -ForegroundColor Green
    Write-Host " 1) Levantar stack completo (build + up -d)"
    Write-Host " 2) Levantar solo backend"
    Write-Host " 3) Levantar solo frontend"
    Write-Host " 4) Reiniciar un servicio (restart, sin rebuild)"
    Write-Host " 5) Parar el stack (down)"
    Write-Host " 6) Ver logs de un servicio (tail -f)"
    Write-Host " 7) Ver estado de los contenedores (ps)"
    Write-Host " 8) Aplicar migraciones pendientes (goose up)"
    Write-Host " 9) Revertir la ultima migracion (goose down)"
    Write-Host "10) Ver estado de las migraciones (goose status)"
    Write-Host "11) Tests del backend (go test ./...)"
    Write-Host "12) Regenerar swagger"
    Write-Host "13) Regenerar sqlc"
    Write-Host " 0) Salir"
    Write-Host ""
}

function Read-Service {
    param([string]$Prompt = "Servicio (backend/frontend/postgres/redis)")
    return Read-Host $Prompt
}

function Run-Action {
    param([string]$Choice)
    switch ($Choice) {
        '1'  { Action-UpAll }
        '2'  { Action-UpService -Service 'backend' }
        '3'  { Action-UpService -Service 'frontend' }
        '4'  { Action-Restart -Service (Read-Service) }
        '5'  { Action-Down }
        '6'  { Action-Logs -Service (Read-Service) }
        '7'  { Action-Ps }
        '8'  { Action-MigrateUp }
        '9'  { Action-MigrateDown }
        '10' { Action-MigrateStatus }
        '11' { Action-BackendTest }
        '12' { Action-Swagger }
        '13' { Action-Sqlc }
        '0'  { return $false }
        default { Write-Host "Opcion no reconocida." -ForegroundColor Yellow }
    }
    return $true
}

if ($Once) {
    Run-Action -Choice $Once | Out-Null
    exit 0
}

$keepGoing = $true
while ($keepGoing) {
    Show-Menu
    $choice = Read-Host "Elige una opcion"
    $keepGoing = Run-Action -Choice $choice
}
