#!/bin/bash

# Get the latest job ID and increment it
LATEST_ID=$(curl -s -X GET http://data.triggerx.network:8080/api/jobs/latest-id | jq -r '.latest_jobID')
NEW_JOB_ID=$((LATEST_ID + 1))
echo "Latest job ID: $LATEST_ID"
echo "New job ID: $NEW_JOB_ID"
# Array of user addresses
USER_ADDRESSES=(
    "0x8A3bEcE42E6C56A96C6D69537e784D88401C8b9F"
    "0x2F5e2E9C62F2A0b9C95D56C43e8A3B075f5A4e1D"
    "0xB1c4D2f8E69A1c5d4a52C8bB962b9E4b2F8D5e3A"
    "0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9E"
)

# Get random user address
RANDOM_INDEX=$((RANDOM % ${#USER_ADDRESSES[@]}))
SELECTED_USER_ADDRESS=${USER_ADDRESSES[$RANDOM_INDEX]}

GAS_PRICE=4

CHAIN_ID=11155420

JOB_TYPE=(1 2 3)
RANDOM_JOB_TYPE_INDEX=$((RANDOM % ${#JOB_TYPE[@]}))
SELECTED_JOB_TYPE=${JOB_TYPE[$RANDOM_JOB_TYPE_INDEX]}

# Array of job cost predictions
JOB_COST_PREDICTIONS=(208 219 256 303)

# Get random job cost prediction
RANDOM_COST_INDEX=$((RANDOM % ${#JOB_COST_PREDICTIONS[@]}))
SELECTED_JOB_COST=${JOB_COST_PREDICTIONS[$RANDOM_COST_INDEX]}

# Create a new job with the incremented ID
curl -X POST http://data.triggerx.network:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d "{
    \"userAddress\": \"$SELECTED_USER_ADDRESS\",
    \"stakeAmount\": $((GAS_PRICE * SELECTED_JOB_COST)),
    \"jobID\": $NEW_JOB_ID,
    \"jobType\": $SELECTED_JOB_TYPE,
    \"chainID\": $CHAIN_ID,
    \"timeFrame\": 5,
    \"timeInterval\": 10,
    \"triggerContractAddress\": \"0xF1d505d1f6df11795c77A8A1b7476609E7b6361a\",
    \"triggerEvent\": \"Staked(address indexed user, uint256 amount)\",
    \"targetContractAddress\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\",
    \"targetFunction\": \"execute\",
    \"argType\": 1,
    \"arguments\": [\"19\", \"91\"],
    \"recurring\": true,
    \"jobCostPrediction\": $SELECTED_JOB_COST,
    \"scriptFunction\": \"checker\",
    \"scriptIPFSUrl\": \"https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a\",
    \"priority\": 1,
    \"security\": 1,
    \"linkJobID\": 0
}"

