#!/bin/bash

echo "Testing ScyllaDB multi-node replication..."

# Insert data through node 1
echo "Inserting test data into node 1..."
docker exec -i triggerx-scylla-1 cqlsh -e "USE triggerx; INSERT INTO user_data (user_id, user_address, created_at) VALUES (999, 'test_replication', toTimestamp(now()));"

# Read data from node 2
echo "Reading test data from node 2..."
docker exec -i triggerx-scylla-2 cqlsh -e "USE triggerx; SELECT * FROM user_data WHERE user_id = 999;"

# Test failover by stopping node 1
echo "Testing failover by stopping node 1..."
docker stop triggerx-scylla-1

echo "Waiting for node to stop..."
sleep 5

# Try to read data from node 2
echo "Reading data from node 2 after node 1 is down..."
docker exec -i triggerx-scylla-2 cqlsh -e "USE triggerx; SELECT * FROM user_data WHERE user_id = 999;"

# Restart node 1
echo "Restarting node 1..."
docker start triggerx-scylla-1

echo "Waiting for node 1 to rejoin..."
sleep 10

# Check cluster status
echo "Checking cluster status after restart..."
docker exec -i triggerx-scylla-2 nodetool status

echo "Test complete." 