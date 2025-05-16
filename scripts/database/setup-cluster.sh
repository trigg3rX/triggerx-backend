#!/bin/bash

# Function to cleanup Docker resources
cleanup() {
    echo "Cleaning up existing resources..."
    docker compose down -v
    docker network prune -f
}

# Function to check if ScyllaDB is ready
check_scylla_ready() {
    local container=$1
    echo "Checking $container status..."
    
    # Check if container is running
    if ! docker ps | grep -q $container; then
        echo "$container is not running"
        docker logs $container
        return 1
    fi

    # Try to connect using cqlsh
    if ! docker exec -i $container cqlsh -e "SELECT now() FROM system.local" >/dev/null 2>&1; then
        echo "$container is not accepting connections"
        docker exec $container nodetool status || true
        return 1
    fi

    return 0
}

# Cleanup first
cleanup

# Start the ScyllaDB cluster
echo "Starting ScyllaDB cluster..."
docker compose up -d

echo "Waiting for ScyllaDB nodes to become available..."
echo "This might take a minute or two for first initialization..."

# Initial delay to let the containers start properly
sleep 30

# Wait for first node to be ready
max_attempts=30
attempt=0
while ! check_scylla_ready "triggerx-scylla-1" && [ $attempt -lt $max_attempts ]; do
    attempt=$(( $attempt + 1 ))
    echo "Waiting for ScyllaDB node 1 to be ready... Attempt $attempt/$max_attempts"
    sleep 10
done

if [ $attempt -eq $max_attempts ]; then
    echo "Timed out waiting for ScyllaDB node 1 to be ready"
    docker logs triggerx-scylla-1
    cleanup
    exit 1
fi

echo "Node 1 is ready! Waiting for node 2..."

# Wait for second node to be ready
attempt=0
while ! check_scylla_ready "triggerx-scylla-2" && [ $attempt -lt $max_attempts ]; do
    attempt=$(( $attempt + 1 ))
    echo "Waiting for ScyllaDB node 2 to be ready... Attempt $attempt/$max_attempts"
    sleep 10
done

if [ $attempt -eq $max_attempts ]; then
    echo "Timed out waiting for ScyllaDB node 2 to be ready"
    docker logs triggerx-scylla-2
    cleanup
    exit 1
fi

echo "Both nodes are ready! Checking cluster status..."
docker exec triggerx-scylla-1 nodetool status

# Wait for both nodes to join cluster
max_attempts=10
attempt=0
while [ $attempt -lt $max_attempts ]; do
    node_count=$(docker exec triggerx-scylla-1 nodetool status | grep "^UN" | wc -l)
    if [ "$node_count" -eq "2" ]; then
        echo "Both nodes are up and running in the cluster!"
        break
    fi
    attempt=$(( $attempt + 1 ))
    if [ $attempt -eq $max_attempts ]; then
        echo "Timed out waiting for nodes to join cluster"
        docker exec triggerx-scylla-1 nodetool status
        cleanup
        exit 1
    fi
    echo "Waiting for both nodes to join cluster... Attempt $attempt/$max_attempts"
    sleep 10
done

# Create keyspace with proper replication
echo "Creating keyspace with NetworkTopologyStrategy..."
if ! docker exec -i triggerx-scylla-1 cqlsh -e "CREATE KEYSPACE IF NOT EXISTS triggerx WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 2};" ; then
    echo "Failed to create keyspace"
    cleanup
    exit 1
fi

# Initialize the database schema
echo "Initializing database schema..."
if ! docker exec -i triggerx-scylla-1 cqlsh < scripts/database/init-db.cql ; then
    echo "Failed to initialize database schema"
    cleanup
    exit 1
fi

# Verify replication
echo "Verifying keyspace replication strategy..."
docker exec -i triggerx-scylla-1 cqlsh -e "DESCRIBE KEYSPACE triggerx;"

# Run repair to ensure data is replicated
echo "Running repair to sync data between nodes..."
docker exec -i triggerx-scylla-1 nodetool repair -pr triggerx || true

echo "ScyllaDB cluster setup complete!" 