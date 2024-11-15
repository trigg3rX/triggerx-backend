#!/bin/bash

# Wait for ScyllaDB to be ready
echo "Waiting for ScyllaDB to be ready..."
sleep 10

# Execute the CQL script
echo "Initializing database..."
docker exec -it triggerx-scylla cqlsh < scripts/init-db.cql

echo "Database initialization completed!" 