#!/bin/bash

# This script tests the job creation API with node failures

# Set the API URL
API_URL="http://localhost:9002/api/jobs"  # Corrected URL with /api prefix

# Function to create a job via API
create_job() {
  local job_id=$1
  echo "Creating job with ID parameter: $job_id"
  
  # Create a job with a unique identifier (using script_ipfs_url to make it unique)
  curl -X POST $API_URL \
    -H "Content-Type: application/json" \
    -d "[{
      \"user_address\": \"0x1234567890abcdef1234567890abcdef12345$job_id\",
      \"stake_amount\": 1000000000000000000,
      \"token_amount\": 1000000000000000000,
      \"task_definition_id\": 1,
      \"priority\": 1,
      \"security\": 1,
      \"time_frame\": 3600,
      \"recurring\": false,
      \"time_interval\": 300,
      \"trigger_chain_id\": \"1\",
      \"trigger_contract_address\": \"0x1234567890abcdef1234567890abcdef12345678\",
      \"trigger_event\": \"Transfer(address,address,uint256)\",
      \"script_ipfs_url\": \"ipfs://Qm123456789_$job_id\",
      \"script_trigger_function\": \"trigger\",
      \"target_chain_id\": \"1\",
      \"target_contract_address\": \"0x1234567890abcdef1234567890abcdef12345678\",
      \"target_function\": \"execute\",
      \"arg_type\": 0,
      \"arguments\": [\"arg1\", \"arg2\"],
      \"script_target_function\": \"execute\"
    }]" 
  
  echo -e "\n"
}

# Function to check if a job exists (now we need to check by user address)
check_job() {
  local user_address="0x1234567890abcdef1234567890abcdef12345$1"
  echo "Checking job for user address: $user_address via node $2"
  
  docker exec -i triggerx-scylla-$2 cqlsh -e "USE triggerx; SELECT job_id, user_id FROM job_data WHERE user_id IN (SELECT user_id FROM user_data WHERE user_address = '$user_address' ALLOW FILTERING) ALLOW FILTERING;"
  echo -e "\n"
}

# Start the API server (if not already running)
echo "Make sure your API server is running. Press Enter to continue..."
read

# Test 1: Create a job with both nodes up
echo "=== Test 1: Create job with both nodes up ==="
create_job 1001
check_job 1001 1
check_job 1001 2

# Test 2: Stop node 1 and create a job
echo "=== Test 2: Create job with node 1 down ==="
echo "Stopping node 1..."
docker stop triggerx-scylla-1
sleep 5
create_job 1002
check_job 1002 2

# Test 3: Restart node 1 and check if it syncs
echo "=== Test 3: Check if node 1 syncs after restart ==="
echo "Restarting node 1..."
docker start triggerx-scylla-1
sleep 10
check_job 1002 1

# Test 4: Stop node 2 and create a job
echo "=== Test 4: Create job with node 2 down ==="
echo "Stopping node 2..."
docker stop triggerx-scylla-2
sleep 5
create_job 1003
check_job 1003 1

# Test 5: Restart node 2 and check if it syncs
echo "=== Test 5: Check if node 2 syncs after restart ==="
echo "Restarting node 2..."
docker start triggerx-scylla-2
sleep 10
check_job 1003 2

# Test 6: Check cluster status
echo "=== Test 6: Check final cluster status ==="
docker exec -i triggerx-scylla-1 nodetool status

echo "Tests completed!" 