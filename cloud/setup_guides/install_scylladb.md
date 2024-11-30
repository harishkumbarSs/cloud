# Installing ScyllaDB on Windows

Since ScyllaDB doesn't have native Windows support, we'll use Docker to run it:

1. Install Docker Desktop for Windows if you haven't already:
   - Download from: https://www.docker.com/products/docker-desktop
   - Install and follow the setup instructions
   - Start Docker Desktop

2. Open PowerShell and run these commands:

```powershell
# Create a Docker network for ScyllaDB
docker network create scylla-network

# Pull and run ScyllaDB container
docker run --name scylla-node -d `
  --network scylla-network `
  -p 9042:9042 `
  scylladb/scylla

# Wait for ScyllaDB to start (about 30 seconds)
# You can check logs with:
docker logs scylla-node

# Initialize the database with our schema
# First, copy schema.cql into the container
docker cp ./internal/db/schema.cql scylla-node:/schema.cql

# Then run the schema
docker exec -it scylla-node cqlsh -f /schema.cql
```

To stop ScyllaDB:
```powershell
docker stop scylla-node
```

To start it again:
```powershell
docker start scylla-node
```
