#!/bin/bash

if [ $# -ne 1 ]; then
  echo "Usage: $0 <endpoint_type (1-4)>"
  echo "1: Deploy Chain (POST /orbit-chain/deploy)"
  echo "2: Get User Chains (GET /orbit-chain/user/:user_address)"
  echo "3: Get Chain Status (GET /orbit-chain/:chain_id/status)"
  echo "4: Get All Chains Dashboard (GET /orbit-chain/dashboard)"
  exit 1
fi

ENDPOINT_TYPE=$1

if ! [[ "$ENDPOINT_TYPE" =~ ^[1-4]$ ]]; then
  echo "Error: endpoint_type must be an integer between 1 and 4."
  exit 1
fi

# Base URL - change this to your server URL
BASE_URL="http://localhost:9002/api/orbit-chain"

# API Key - change this to your API key
API_KEY="ADMIN"

case $ENDPOINT_TYPE in
  1)
    echo "Deploying new Orbit chain..."
    
    # Generate unique chain ID using timestamp
    CHAIN_ID=$(date +%s)
    
    # Generate chain name
    adj=$(shuf -n 1 scripts/adjectives.txt 2>/dev/null || echo "Amazing")
    noun=$(shuf -n 1 scripts/nouns.txt 2>/dev/null || echo "Chain")
    CHAIN_NAME="$adj $noun"
    
    # Default addresses - change these as needed
    USER_ADDRESS="0x7Db951c0E6D8906687B459427eA3F3F2b456473B"
    OWNER_ADDRESS="0x7Db951c0E6D8906687B459427eA3F3F2b456473B"
    BATCH_POSTER="0x7Db951c0E6D8906687B459427eA3F3F2b456473B"
    VALIDATOR="0x7Db951c0E6D8906687B459427eA3F3F2b456473B"
    
    echo "Chain ID: $CHAIN_ID"
    echo "Chain Name: $CHAIN_NAME"
    echo "User Address: $USER_ADDRESS"
    
    curl -X POST "$BASE_URL/deploy" \
      -H "Content-Type: application/json" \
      -H "X-API-KEY: $API_KEY" \
      -d "{
        \"chain_id\": $CHAIN_ID,
        \"chain_name\": \"$CHAIN_NAME\",
        \"owner_address\": \"$OWNER_ADDRESS\",
        \"batch_poster\": \"$BATCH_POSTER\",
        \"validator\": \"$VALIDATOR\",
        \"user_address\": \"$USER_ADDRESS\",
        \"native_token\": \"\",
        \"token_name\": \"\",
        \"token_symbol\": \"\",
        \"token_decimals\": 0,
        \"max_data_size\": 0,
        \"max_fee_per_gas_for_retryables\": \"\"
      }"
    ;;
    
  2)
    echo "Getting chains for user..."
    
    # Default user address - change this as needed
    USER_ADDRESS="0x7Db951c0E6D8906687B459427eA3F3F2b456473B"
    
    echo "User Address: $USER_ADDRESS"
    
    curl -X GET "$BASE_URL/user/$USER_ADDRESS" \
      -H "Content-Type: application/json" \
      -H "X-API-KEY: $API_KEY"
    ;;
    
  3)
    echo "Getting chain deployment status..."
    
    # You can change this chain ID to query a specific chain
    CHAIN_ID="1234567890"
    
    echo "Chain ID: $CHAIN_ID"
    
    curl -X GET "$BASE_URL/$CHAIN_ID/status" \
      -H "Content-Type: application/json" \
      -H "X-API-KEY: $API_KEY"
    ;;
    
  4)
    echo "Getting all chains dashboard..."
    
    curl -X GET "$BASE_URL/dashboard" \
      -H "Content-Type: application/json" \
      -H "X-API-KEY: $API_KEY"
    ;;
esac

echo ""
echo "Request completed."
