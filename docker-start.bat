@echo off
REM Food Order Tracking - Docker Deployment Script
REM For Windows Command Prompt
REM Usage: docker-start up
REM        docker-start down
REM        docker-start logs

setlocal enabledelayedexpansion

set COMMAND=%1
if "%COMMAND%"=="" set COMMAND=up

echo.
echo ╔════════════════════════════════════════════════════════════════╗
echo ║  Food Order Tracking - Docker Deployment                      ║
echo ╚════════════════════════════════════════════════════════════════╝
echo.

REM Check if Docker is running
echo Checking Docker Desktop...
docker ps >nul 2>&1
if errorlevel 1 (
    echo ✗ Docker Desktop is not running
    echo Please start Docker Desktop and try again
    pause
    exit /b 1
)
echo ✓ Docker is running
echo.

REM Execute command
if "%COMMAND%"=="up" (
    echo Starting application...
    docker-compose up -d
    if !errorlevel! equ 0 (
        echo.
        echo ✓ Application started successfully!
        echo.
        echo Access the application:
        echo   Frontend: http://localhost
        echo   API:      http://localhost:8080/api
        echo   Database: localhost:5432 (postgres/postgres^)
        echo.
        echo Useful commands:
        echo   docker-compose ps              - Show running containers
        echo   docker-compose logs -f         - View logs
        echo   docker-compose restart backend - Restart backend
        echo   docker-compose down            - Stop all services
        echo.
    )
) else if "%COMMAND%"=="down" (
    echo Stopping application...
    docker-compose down
    echo ✓ Application stopped
) else if "%COMMAND%"=="logs" (
    echo Showing logs (Ctrl+C to exit)...
    docker-compose logs -f
) else if "%COMMAND%"=="restart" (
    echo Restarting application...
    docker-compose restart
    echo ✓ Application restarted
) else if "%COMMAND%"=="build" (
    echo Building images...
    docker-compose build --no-cache
    echo ✓ Build complete
) else if "%COMMAND%"=="clean" (
    echo Cleaning up Docker resources...
    echo This will remove containers, volumes, and database data
    set /p confirm="Continue? (y/N): "
    if "!confirm!"=="y" (
        docker-compose down -v
        echo ✓ Cleanup complete
    ) else (
        echo Cancelled
    )
) else if "%COMMAND%"=="status" (
    echo Container Status:
    docker-compose ps
) else if "%COMMAND%"=="help" (
    echo Usage: docker-start [command]
    echo.
    echo Commands:
    echo   up       Start application (default^)
    echo   down     Stop application
    echo   logs     View application logs
    echo   restart  Restart application
    echo   build    Build Docker images
    echo   clean    Remove containers ^& database (WARNING^)
    echo   status   Show container status
    echo   help     Show this help message
    echo.
    echo Examples:
    echo   docker-start              - Start in background
    echo   docker-start logs         - Show live logs
    echo   docker-start down         - Stop services
) else (
    echo Unknown command: %COMMAND%
    echo Use 'help' for available commands
    exit /b 1
)

endlocal
