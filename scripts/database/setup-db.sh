#!/bin/bash

# Execute the CQL script
echo "Initializing database schema..."
docker exec -i triggerx-scylla cqlsh < scripts/database/init-db.cql

# Check if the keyspace was created
if [ $? -eq 0 ]; then
    echo "Database schema initialized successfully"
else
    echo "Failed to initialize database schema"
    exit 1
fi
