@echo off
echo Setting up Cloud Storage Application...

:: Check if Docker is installed
docker --version > nul 2>&1
if %errorlevel% neq 0 (
    echo Docker is not installed! Please install Docker Desktop from:
    echo https://www.docker.com/products/docker-desktop
    exit /b 1
)

:: Create Docker network
echo Creating Docker network...
docker network create cloud-network 2>nul

:: Start ScyllaDB
echo Starting ScyllaDB...
docker run --name scylla-node -d ^
    --network cloud-network ^
    -p 9042:9042 ^
    scylladb/scylla

:: Start MinIO
echo Starting MinIO...
docker run --name minio-node -d ^
    --network cloud-network ^
    -p 9000:9000 ^
    -p 9001:9001 ^
    -e "MINIO_ROOT_USER=minioadmin" ^
    -e "MINIO_ROOT_PASSWORD=minioadmin" ^
    minio/minio server /data --console-address ":9001"

:: Wait for services to start
echo Waiting for services to initialize (30 seconds)...
timeout /t 30 /nobreak

:: Initialize ScyllaDB schema
echo Initializing database schema...
docker cp ./internal/db/schema.cql scylla-node:/schema.cql
docker exec -it scylla-node cqlsh -f /schema.cql

:: Install Go dependencies
echo Installing Go dependencies...
go get github.com/gocql/gocql
go get github.com/minio/minio-go/v7
go mod tidy

echo Setup complete! Services are running:
echo - ScyllaDB: localhost:9042
echo - MinIO: localhost:9000 (API) and localhost:9001 (Console)
echo.
echo You can access the MinIO Console at: http://localhost:9001
echo Username: minioadmin
echo Password: minioadmin
