#!/bin/bash

if [ $# -ne 1 ]; then
  echo "Usage: $0 <job_type (1-6)>"
  exit 1
fi

PRIVATE_KEY=
CHAIN_ID=421614
JOB_REGISTRY_CONTRACT_ADDRESS=0x476ACc7949a95e31144cC84b8F6BC7abF0967E4b
TEST_CONTRACT_ADDRESS=0xa92f95FDeF3DB6B2aA115548376c7a2429711497
ALCHEMY_API_KEY=
RPC_URL=https://arb-sepolia.g.alchemy.com/v2/$ALCHEMY_API_KEY
IPFS_URL=https://teal-random-koala-993.mypinata.cloud/ipfs/bafkreif426p7t7takzhw3g6we2h6wsvf27p5jxj3gaiynqf22p3jvhx4la

read -r -d '' TEST_EVENT_ABI <<'EOF'
[
  {
    "anonymous": false,
    "inputs": [
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "previousValue",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "newValue",
        "type": "uint256"
      },
      {
        "indexed": false,
        "internalType": "uint256",
        "name": "incrementAmount",
        "type": "uint256"
      }
    ],
    "name": "CounterIncremented",
    "type": "event"
  }
]
EOF

read -r -d '' TEST_FUNCTION_ABI <<'EOF'
[
  {
    "inputs": [
      { "internalType": "uint256", "name": "amount", "type": "uint256" }
    ],
    "name": "incrementBy",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]
EOF

TASK_DEFINITION_ID=$1

if ! [[ "$TASK_DEFINITION_ID" =~ ^[1-6]$ ]]; then
  echo "Error: job_type must be an integer between 1 and 6."
  exit 1
fi

adj=$(shuf -n 1 scripts/adjectives.txt)
noun=$(shuf -n 1 scripts/nouns.txt)

JOB_TITLE="$adj $noun"
RECURRING=false
JOB_DATA=""
DYNAMIC_ARGUMENTS_SCRIPT_URL=""
JOB_DATA=$(printf '0x%064x' 1)

case $TASK_DEFINITION_ID in
  1)
    TASK_DEFINITION_ID=1 # Time Based, Static Args
    TIME_FRAME=35
    TIME_INTERVAL=32
    ARG_TYPE=1
    echo "Creating Time-based Static Args Job..."
    ;;
  2)
    TASK_DEFINITION_ID=2 # Time Based, Dynamic Args
    TIME_FRAME=35
    TIME_INTERVAL=32
    ARG_TYPE=2
    DYNAMIC_ARGUMENTS_SCRIPT_URL=$IPFS_URL
    echo "Creating Time-based Dynamic Args Job..."
    ;;
  3)
    TASK_DEFINITION_ID=3 # Event Based, Static Args
    TIME_FRAME=2500
    TIME_INTERVAL=0
    ARG_TYPE=1
    echo "Creating Event-based Static Args Job..."
    ;;
  4)
    TASK_DEFINITION_ID=4 # Event Based, Dynamic Args
    TIME_FRAME=2500
    TIME_INTERVAL=0
    ARG_TYPE=2
    DYNAMIC_ARGUMENTS_SCRIPT_URL=$IPFS_URL
    echo "Creating Event-based Dynamic Args Job..."
    ;;
  5)
    TASK_DEFINITION_ID=5 # Condition Based, Static Args
    TIME_FRAME=20
    TIME_INTERVAL=0
    ARG_TYPE=1
    echo "Creating Condition-based Static Args Job..."
    ;;
  6)
    TASK_DEFINITION_ID=6 # Condition Based, Dynamic Args
    TIME_FRAME=20
    TIME_INTERVAL=0
    ARG_TYPE=2
    DYNAMIC_ARGUMENTS_SCRIPT_URL=$IPFS_URL
    echo "Creating Condition-based Dynamic Args Job..."
    ;;
esac

if ! command -v cast >/dev/null 2>&1; then
  echo "Error: foundry 'cast' CLI is required"
  exit 1
fi

if [ -n "$DYNAMIC_ARGUMENTS_SCRIPT_URL" ]; then
  IPFS_HASH_BYTES32=$(cast keccak "$DYNAMIC_ARGUMENTS_SCRIPT_URL")
  if [ $? -ne 0 ] || [ -z "$IPFS_HASH_BYTES32" ]; then
    echo "Error: failed to hash dynamic arguments script URL"
    exit 1
  fi
else
  IPFS_HASH_BYTES32=""
fi

case $TASK_DEFINITION_ID in
  1)
    JOB_DATA=$(cast abi-encode "encode(uint256)" $TIME_INTERVAL)
    ;;
  2)
    if [ -z "$IPFS_HASH_BYTES32" ]; then
      echo "Error: dynamic arguments script URL required for job type 2"
      exit 1
    fi
    JOB_DATA=$(cast abi-encode "encode(uint256,bytes32)" $TIME_INTERVAL $IPFS_HASH_BYTES32)
    ;;
  3|5)
    JOB_DATA=$(cast abi-encode "encode(bool)" $RECURRING)
    ;;
  4|6)
    if [ -z "$IPFS_HASH_BYTES32" ]; then
      echo "Error: dynamic arguments script URL required for job type $TASK_DEFINITION_ID"
      exit 1
    fi
    JOB_DATA=$(cast abi-encode "encode(bool,bytes32)" $RECURRING $IPFS_HASH_BYTES32)
    ;;
  *)
    JOB_DATA=0x
    ;;
esac

if [ $? -ne 0 ] || [ -z "$JOB_DATA" ]; then
  echo "Error: failed to encode job data"
  exit 1
fi

echo "\nCalling CreateJob() to TriggerXJobRegistry..."

EVENT_SIGNATURE=0x737fc62fbb05dd9fb7c799e68796a2d7c8324e310af7b656c0580e7b3cf8bf8a

echo "Submitting transaction..."

CAST_RESULT=$(cast send \
  --chain $CHAIN_ID \
  --rpc-url $RPC_URL \
  --private-key $PRIVATE_KEY \
  --json \
  $JOB_REGISTRY_CONTRACT_ADDRESS \
  "createJob(string,uint8,uint256,address,bytes)" "$JOB_TITLE" $TASK_DEFINITION_ID $TIME_FRAME $TEST_CONTRACT_ADDRESS $JOB_DATA)

if [ $? -ne 0 ]; then
  echo "Error: failed to submit transaction"
  exit 1
fi

TX_HASH=$(echo "$CAST_RESULT" | jq -r '.transactionHash // empty')

if [ -z "$TX_HASH" ]; then
  echo "Error: unable to extract transaction hash"
  exit 1
fi

echo "Transaction hash: $TX_HASH"

JOB_ID_HEX=$(echo "$CAST_RESULT" | jq -r --arg sig "$EVENT_SIGNATURE" '.logs[] | select((.topics[0] | ascii_downcase) == ($sig | ascii_downcase)) | .topics[1]' | head -n 1)

if [ -z "$JOB_ID_HEX" ]; then
  RECEIPT=$(cast receipt "$TX_HASH" --rpc-url $RPC_URL --json)
  JOB_ID_HEX=$(echo "$RECEIPT" | jq -r --arg sig "$EVENT_SIGNATURE" '.logs[] | select((.topics[0] | ascii_downcase) == ($sig | ascii_downcase)) | .topics[1]' | head -n 1)
fi

if [ -z "$JOB_ID_HEX" ]; then
  echo "Error: unable to extract JobCreated event from logs"
  exit 1
fi

JOB_ID=$(cast --to-dec "$JOB_ID_HEX")

echo "Job created with ID: $JOB_ID"

sleep 3

curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: ADMIN" \
  -d "[
    {
      \"user_address\": \"0x7db951c0e6d8906687b459427ea3f3f2b456473b\",
      \"ether_balance\": 50000000000000000,
      \"token_balance\": 50000000000000000000,
      \"created_chain_id\": \"$CHAIN_ID\",
      \"job_id\": \"$JOB_ID\",
      \"job_title\": \"$JOB_TITLE\",
      \"task_definition_id\": $TASK_DEFINITION_ID,
      \"custom\": true,
      \"time_frame\": $TIME_FRAME,
      \"recurring\": false,
      \"job_cost_prediction\": 0.1,
      \"timezone\": \"Asia/Calcutta\",
      \"schedule_type\": \"interval\",
      \"time_interval\": $TIME_INTERVAL,
      \"cron_expression\": \"0 0 * * *\",
      \"specific_schedule\": \"2025-01-01 00:00:00\",
      \"trigger_chain_id\": \"421614\",
      \"trigger_contract_address\": \"$TEST_CONTRACT_ADDRESS\",
      \"trigger_event\": \"Transfer(address,address,uint256)\",
      \"event_filter_para_name\": \"to\",
      \"event_filter_value\": \"0xC9dC9c361c248fFA0890d7E1a263247670914980\",
      \"condition_type\": \"less_than\",
      \"upper_limit\": 92,
      \"lower_limit\": 89,
      \"value_source_type\": \"api\",
      \"value_source_url\": \"http://localhost:8080/price\",
      \"target_chain_id\": \"421614\",
      \"target_contract_address\": \"$TEST_CONTRACT_ADDRESS\",
      \"target_function\": \"implementation\",
      \"abi\": \"[{\\\"inputs\\\":[],\\\"name\\\":\\\"implementation\\\",\\\"outputs\\\":[{\\\"internalType\\\":\\\"address\\\",\\\"name\\\":\\\"implementation_\\\",\\\"type\\\":\\\"address\\\"}],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
      \"arg_type\": $ARG_TYPE,
      \"is_imua\": false,
      \"arguments\": [\"3\"],
      \"created_chain_id\": \"$CHAIN_ID\",
      \"dynamic_arguments_script_url\": \"$DYNAMIC_ARGUMENTS_SCRIPT_URL\"
    }
  ]"

echo "\n"

if [ $TASK_DEFINITION_ID -eq 3 ] || [ $TASK_DEFINITION_ID -eq 4 ]; then
  sleep 5 
  echo "\nCalling increment() to trigger the event..."

  cast send \
    --async \
    --chain $CHAIN_ID \
    --rpc-url https://arb-sepolia.g.alchemy.com/v2/$ALCHEMY_API_KEY \
    --private-key $PRIVATE_KEY \
    $TEST_CONTRACT_ADDRESS "increment()" \
    -- --broadcast
fi

echo "\n"
