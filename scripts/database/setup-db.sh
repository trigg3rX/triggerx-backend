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

sleep 3
echo "Inserting test data into Database..."
docker exec -i triggerx-scylla cqlsh < scripts/database/test-data.cql

# Check if the test data was inserted
if [ $? -eq 0 ]; then
    echo "Test data inserted successfully"
else
    echo "Failed to insert test data"
    exit 1
fi