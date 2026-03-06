# Food Order Tracking - Docker Deployment Script
# For Windows PowerShell
# Usage: .\docker-start.ps1

param(
    [string]$Command = "up",
    [switch]$Rebuild = $false,
    [switch]$Detached = $true
)

$ErrorActionPreference = "Stop"

Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  Food Order Tracking - Docker Deployment                      ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
Write-Host "Checking Docker Desktop..." -ForegroundColor Yellow
try {
    $null = docker ps 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Docker not running"
    }
    Write-Host "✓ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker Desktop is not running" -ForegroundColor Red
    Write-Host "Please start Docker Desktop and try again" -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# Execute command
switch ($Command.ToLower()) {
    "up" {
        Write-Host "Starting application..." -ForegroundColor Yellow
        if ($Rebuild) {
            Write-Host "Building images..." -ForegroundColor Yellow
            docker-compose build --no-cache
        }
        
        $args = if ($Detached) { @("-d") } else { @() }
        docker-compose up @args
        
        if ($Detached) {
            Write-Host ""
            Write-Host "✓ Application started successfully!" -ForegroundColor Green
            Write-Host ""
            Write-Host "Access the application:" -ForegroundColor Cyan
            Write-Host "  Frontend: http://localhost" -ForegroundColor White
            Write-Host "  API:      http://localhost:8080/api" -ForegroundColor White
            Write-Host "  Database: localhost:5432 (postgres/postgres)" -ForegroundColor White
            Write-Host ""
            Write-Host "Useful commands:" -ForegroundColor Cyan
            Write-Host "  docker-compose ps              # Show running containers" -ForegroundColor White
            Write-Host "  docker-compose logs -f         # View logs" -ForegroundColor White
            Write-Host "  docker-compose restart backend # Restart backend" -ForegroundColor White
            Write-Host "  docker-compose down            # Stop all services" -ForegroundColor White
            Write-Host ""
        }
    }
    "down" {
        Write-Host "Stopping application..." -ForegroundColor Yellow
        docker-compose down
        Write-Host "✓ Application stopped" -ForegroundColor Green
    }
    "logs" {
        Write-Host "Showing logs (Ctrl+C to exit)..." -ForegroundColor Yellow
        docker-compose logs -f
    }
    "restart" {
        Write-Host "Restarting application..." -ForegroundColor Yellow
        docker-compose restart
        Write-Host "✓ Application restarted" -ForegroundColor Green
    }
    "build" {
        Write-Host "Building images..." -ForegroundColor Yellow
        docker-compose build --no-cache
        Write-Host "✓ Build complete" -ForegroundColor Green
    }
    "clean" {
        Write-Host "Cleaning up Docker resources..." -ForegroundColor Yellow
        Write-Host "This will remove containers, volumes, and database data" -ForegroundColor Red
        $confirm = Read-Host "Continue? (y/N)"
        if ($confirm -eq "y") {
            docker-compose down -v
            Write-Host "✓ Cleanup complete" -ForegroundColor Green
        } else {
            Write-Host "Cancelled" -ForegroundColor Yellow
        }
    }
    "status" {
        Write-Host "Container Status:" -ForegroundColor Cyan
        docker-compose ps
    }
    "help" {
        Write-Host "Usage: .\docker-start.ps1 [command] [options]" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Commands:" -ForegroundColor Yellow
        Write-Host "  up       Start application (default)" -ForegroundColor White
        Write-Host "  down     Stop application" -ForegroundColor White
        Write-Host "  logs     View application logs" -ForegroundColor White
        Write-Host "  restart  Restart application" -ForegroundColor White
        Write-Host "  build    Build Docker images" -ForegroundColor White
        Write-Host "  clean    Remove containers & database (WARNING)" -ForegroundColor White
        Write-Host "  status   Show container status" -ForegroundColor White
        Write-Host "  help     Show this help message" -ForegroundColor White
        Write-Host ""
        Write-Host "Options:" -ForegroundColor Yellow
        Write-Host "  -Rebuild    Rebuild images before starting" -ForegroundColor White
        Write-Host "  -NoDetach   Run in foreground (see logs)" -ForegroundColor White
        Write-Host ""
        Write-Host "Examples:" -ForegroundColor Green
        Write-Host "  .\docker-start.ps1                      # Start in background" -ForegroundColor White
        Write-Host "  .\docker-start.ps1 -Rebuild             # Rebuild and start" -ForegroundColor White
        Write-Host "  .\docker-start.ps1 up -NoDetach        # Start in foreground" -ForegroundColor White
        Write-Host "  .\docker-start.ps1 logs                 # Show live logs" -ForegroundColor White
        Write-Host "  .\docker-start.ps1 down                 # Stop services" -ForegroundColor White
    }
    default {
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Write-Host "Use 'help' for available commands" -ForegroundColor Yellow
        exit 1
    }
}
