#!/bin/bash

# Start the ScyllaDB cluster
echo "Starting ScyllaDB cluster..."
docker compose down
docker compose up -d

echo "Waiting for ScyllaDB nodes to become available..."

# Function to check if ScyllaDB is ready
check_scylla_ready() {
  docker exec -i triggerx-scylla-1 cqlsh -e "SELECT now() FROM system.local" >/dev/null 2>&1
  return $?
}

# Wait for ScyllaDB to be ready with timeout
max_attempts=30
attempt=0
while ! check_scylla_ready && [ $attempt -lt $max_attempts ]; do
  attempt=$(( $attempt + 1 ))
  echo "Waiting for ScyllaDB to be ready... Attempt $attempt/$max_attempts"
  sleep 5
done

if [ $attempt -eq $max_attempts ]; then
  echo "Timed out waiting for ScyllaDB to be ready"
  exit 1
fi

echo "ScyllaDB is ready!"

# Check node1 status
echo "Checking cluster status..."
docker exec triggerx-scylla-1 nodetool status || true

# Wait for both nodes to join cluster
max_attempts=10
attempt=0
while [ $attempt -lt $max_attempts ]; do
  node_count=$(docker exec triggerx-scylla-1 nodetool status | grep "^UN" | wc -l)
  if [ "$node_count" -eq "2" ]; then
    echo "Both nodes are up and running!"
    break
  fi
  attempt=$(( $attempt + 1 ))
  echo "Waiting for both nodes to join cluster... Attempt $attempt/$max_attempts"
  sleep 10
done

# Create keyspace with proper replication
echo "Creating keyspace with NetworkTopologyStrategy..."
docker exec -i triggerx-scylla-1 cqlsh -e "CREATE KEYSPACE IF NOT EXISTS triggerx WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': 2};"

# Initialize the database schema
echo "Initializing database schema..."
docker exec -i triggerx-scylla-1 cqlsh < scripts/database/init-db.cql

# Verify replication
echo "Verifying keyspace replication strategy..."
docker exec -i triggerx-scylla-1 cqlsh -e "DESCRIBE KEYSPACE triggerx;"

# Run repair to ensure data is replicated
echo "Running repair to sync data between nodes..."
docker exec -i triggerx-scylla-1 nodetool repair -pr || true

echo "ScyllaDB cluster setup complete!" 