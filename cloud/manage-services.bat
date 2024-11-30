@echo off
if "%1"=="" goto usage

if "%1"=="start" (
    echo Starting services...
    docker start scylla-node minio-node
    goto end
)

if "%1"=="stop" (
    echo Stopping services...
    docker stop scylla-node minio-node
    goto end
)

if "%1"=="status" (
    echo Checking service status...
    echo ScyllaDB:
    docker ps -f name=scylla-node --format "{{.Status}}"
    echo.
    echo MinIO:
    docker ps -f name=minio-node --format "{{.Status}}"
    goto end
)

:usage
echo Usage:
echo   %0 start   - Start all services
echo   %0 stop    - Stop all services
echo   %0 status  - Check service status

:end
