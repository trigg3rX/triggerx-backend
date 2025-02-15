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

CHAIN_ID=11155420

# JOB_TYPE=(1 2 3 4)
JOB_TYPE=(3)
RANDOM_JOB_TYPE_INDEX=$((RANDOM % ${#JOB_TYPE[@]}))
SELECTED_JOB_TYPE=${JOB_TYPE[$RANDOM_JOB_TYPE_INDEX]}

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
    \"jobType\": $SELECTED_JOB_TYPE,
    \"user_address\": \"$SELECTED_USER_ADDRESS\",
    \"chain_id\": \"$CHAIN_ID\",
    \"time_frame\": 10000,
    \"time_interval\": 10,
    \"contract_address\": \"0xF1d505d1f6df11795c77A8A1b7476609E7b6361a\",
    \"target_function\": \"Staked(address indexed user, uint256 amount)\",
    \"arg_type\": 1,
    \"arguments\": [\"1000\", \"2000\"],
    \"status\": true,
    \"job_cost_prediction\": $SELECTED_JOB_COST,
    \"script_function\": \"checker\",
    \"script_ipfs_url\": \"https://api.binance.com/api/v3/ticker/price?symbol=ETHUSDT\",
    \"stake_amount\": $((GAS_PRICE * SELECTED_JOB_COST))
}"