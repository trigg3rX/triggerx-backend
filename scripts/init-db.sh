#!/bin/bash

# Wait for ScyllaDB to be ready
echo "Waiting for ScyllaDB to be ready..."
sleep 10

# Execute the CQL script
echo "Initializing database schema..."
docker exec -i triggerx-scylla cqlsh < scripts/init-db.cql

# Check if the keyspace was created
if [ $? -eq 0 ]; then
    echo "Database schema initialized successfully"
else
    echo "Failed to initialize database schema"
    exit 1
fi 