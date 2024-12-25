#!/bin/bash

echo "Seeding database with initial data..."
docker exec -i triggerx-scylla cqlsh < scripts/seed-db.cql

if [ $? -eq 0 ]; then
    echo "Database seeded successfully"
else
    echo "Failed to seed database"
    exit 1
fi
