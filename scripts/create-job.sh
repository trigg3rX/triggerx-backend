#!/bin/bash

# Get the latest job ID and increment it
LATEST_ID=$(curl -s -X GET http://localhost:8080/api/jobs/latest-id | jq -r '.latest_job_id')
NEW_JOB_ID=$((LATEST_ID + 1))

# Array of user addresses
USER_ADDRESSES=(
    "0x7Db951c0E6D8906687B459427eA3F3F2b456473B"
    "0xc073A5E091DC60021058346b10cD5A9b3F0619fE" 
    "0xD5E9061656252a0b44D98C6944B99046FDDf49cA"
    "0xC9dC9c361c248fFA0890d7E1a263247670914980"
)

# Get random user address
RANDOM_INDEX=$((RANDOM % ${#USER_ADDRESSES[@]}))
SELECTED_USER_ADDRESS=${USER_ADDRESSES[$RANDOM_INDEX]}

GAS_PRICE=4

# Array of job cost predictions
JOB_COST_PREDICTIONS=(208 219 256 303)

# Get random job cost prediction
RANDOM_COST_INDEX=$((RANDOM % ${#JOB_COST_PREDICTIONS[@]}))
SELECTED_JOB_COST=${JOB_COST_PREDICTIONS[$RANDOM_COST_INDEX]}

# Create a new job with the incremented ID
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d "{
    \"job_id\": $NEW_JOB_ID,
    \"jobType\": 1,
    \"user_address\": \"$SELECTED_USER_ADDRESS\",
    \"chain_id\": \"11155420\",
    \"time_frame\": 86400,
    \"time_interval\": 10800,
    \"contract_address\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\",
    \"target_function\": \"addTaskId(uint256,uint256)\",
    \"arg_type\": 1,
    \"arguments\": [\"1000\", \"2000\"],
    \"status\": true,
    \"job_cost_prediction\": $SELECTED_JOB_COST,
    \"script_function\": \"checker\",
    \"script_ipfs_url\": \"https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a\",
    \"stake_amount\": $((GAS_PRICE * SELECTED_JOB_COST))
}"