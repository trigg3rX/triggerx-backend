#!/bin/bash

USE_SERVER=false
SHOW_PAYLOAD=false
source .env

TASK_DEFINITION_ID=""

# Parse arguments - handle all orders: [-s] [-o] <id> or <id> [-s] [-o]
for arg in "$@"; do
  case $arg in
    -s)
      USE_SERVER=true
      ;;
    -o)
      SHOW_PAYLOAD=true
      ;;
    [1-9])
      if [ -z "$TASK_DEFINITION_ID" ]; then
        TASK_DEFINITION_ID=$arg
      else
        echo "Error: Multiple job types specified"
        echo "Usage: $0 <job_type (1-9)> [-s] [-o]"
        echo "  -s    Use server (data.triggerx.network) instead of local"
        echo "  -o    Show payload"
        exit 1
      fi
      ;;
    *)
      echo "Error: Invalid argument: $arg"
      echo "Usage: $0 <job_type (1-9)> [-s] [-o]"
      echo "  -s    Use server (data.triggerx.network) instead of local"
      echo "  -o    Show payload"
      exit 1
      ;;
  esac
done

if [ -z "$TASK_DEFINITION_ID" ]; then
  echo "Usage: $0 <job_type (1-9)> [-s] [-o]"
  echo "  -s    Use server (data.triggerx.network) instead of local"
  echo "  -o    Show payload"
  exit 1
fi

# Read from environment variables
if [ -z "$SCRIPT_PRIVATE_KEY" ]; then
  echo "Error: SCRIPT_PRIVATE_KEY environment variable is not set"
  exit 1
fi

if [ -z "$ALCHEMY_API_KEY" ]; then
  echo "Error: ALCHEMY_API_KEY environment variable is not set"
  exit 1
fi

if [ -z "$TRIGGERX_API_KEY" ]; then
  echo "Error: TRIGGERX_API_KEY environment variable is not set"
  exit 1
fi

# Set DB_SERVER_URL based on -s flag
if [ "$USE_SERVER" = true ]; then
  DB_SERVER_URL=https://data.triggerx.network
else
  DB_SERVER_URL=http://localhost:9002
fi

# Derive user address from private key
if ! command -v cast >/dev/null 2>&1; then
  echo "Error: foundry 'cast' CLI is required"
  exit 1
fi

USER_ADDRESS=$(cast wallet address --private-key $SCRIPT_PRIVATE_KEY 2>/dev/null)
if [ $? -ne 0 ] || [ -z "$USER_ADDRESS" ]; then
  echo "Error: failed to derive address from private key"
  exit 1
fi

echo "Derived user address: $USER_ADDRESS"

CHAIN_ID=421614
CREATED_CHAIN_ID="421614"  # String format for API
JOB_REGISTRY_CONTRACT_ADDRESS=0x476ACc7949a95e31144cC84b8F6BC7abF0967E4b
TEST_CONTRACT_ADDRESS=0xa92f95FDeF3DB6B2aA115548376c7a2429711497
RPC_URL=https://arb-sepolia.g.alchemy.com/v2/$ALCHEMY_API_KEY
IPFS_URL=https://teal-random-koala-993.mypinata.cloud/ipfs/bafkreif426p7t7takzhw3g6we2h6wsvf27p5jxj3gaiynqf22p3jvhx4la
CUSTOM_IPFS_URL=https://aqua-tough-swift-909.mypinata.cloud/ipfs/bafkreidvy5nosknw2aqp7dgtdw3cipgxas3fo2ae5q7ttuak5nnoaaluge

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


if ! [[ "$TASK_DEFINITION_ID" =~ ^[1-9]$ ]]; then
  echo "Error: job_type must be an integer between 1 and 9."
  exit 1
fi

adj=$(shuf -n 1 scripts/adjectives.txt)
noun=$(shuf -n 1 scripts/nouns.txt)

JOB_TITLE="$adj $noun"

# Default values for all fields (matching CreateJobData struct)
# USER_ADDRESS is already derived from PRIVATE_KEY above
JOB_TYPE="sdk"
RECURRING=false
JOB_COST_PREDICTION="100"
TIMEZONE="Asia/Calcutta"

IS_SAFE=false
SAFE_ADDRESS=""
SAFE_NAME=""

# Time job fields (defaults)
SCHEDULE_TYPE="interval"
TIME_INTERVAL=0
CRON_EXPRESSION="0 0 * * *"
SPECIFIC_SCHEDULE="2025-01-01 00:00:00"

# Event job fields (defaults)
TRIGGER_CHAIN_ID="421614"
TRIGGER_CONTRACT_ADDRESS=$TEST_CONTRACT_ADDRESS
TRIGGER_EVENT="CounterIncremented(uint256,uint256,uint256)"
EVENT_FILTER_PARA_NAME=""
EVENT_FILTER_VALUE=""

# Condition job fields (defaults)
CONDITION_TYPE="less_than"
UPPER_LIMIT=4000
LOWER_LIMIT=3000
VALUE_SOURCE_TYPE="api"
# VALUE_SOURCE_URL="https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd"
VALUE_SOURCE_URL="http://localhost:8080/api/v1/price"
SELECTED_KEY_ROUTE=""

# Target fields (common for all job types)
TARGET_CHAIN_ID="421614"
TARGET_CONTRACT_ADDRESS=$TEST_CONTRACT_ADDRESS
TARGET_FUNCTION="incrementBy"
ARG_TYPE=1
ARGUMENTS="[\"3\"]"
EXECUTION_SCRIPT_URL=""
EXECUTION_SCRIPT_LANGUAGE="go"
EXECUTION_SCRIPT_HASH=""
MAX_EXECUTION_TIME=50
CHALLENGE_PERIOD=21600

# Task definition specific overrides
case $TASK_DEFINITION_ID in
  1)
    # Time Based, Static Args
    TIME_FRAME=35
    TIME_INTERVAL=32
    ARG_TYPE=1
    echo "Creating Time-based Static Args Job..."
    ;;
  2)
    # Time Based, Dynamic Args
    TIME_FRAME=35
    TIME_INTERVAL=32
    ARG_TYPE=2
    EXECUTION_SCRIPT_LANGUAGE="go"
    EXECUTION_SCRIPT_URL=$IPFS_URL
    echo "Creating Time-based Dynamic Args Job..."
    ;;
  3)
    # Event Based, Static Args
    TIME_FRAME=40
    TIME_INTERVAL=0
    ARG_TYPE=1
    RECURRING=false
    echo "Creating Event-based Static Args Job..."
    ;;
  4)
    # Event Based, Dynamic Args
    TIME_FRAME=40
    TIME_INTERVAL=0
    ARG_TYPE=2
    RECURRING=false
    EXECUTION_SCRIPT_LANGUAGE="go"
    EXECUTION_SCRIPT_URL=$IPFS_URL
    echo "Creating Event-based Dynamic Args Job..."
    ;;
  5)
    # Condition Based, Static Args
    TIME_FRAME=35
    TIME_INTERVAL=0
    ARG_TYPE=1
    RECURRING=false
    echo "Creating Condition-based Static Args Job..."
    ;;
  6)
    # Condition Based, Dynamic Args
    TIME_FRAME=35
    TIME_INTERVAL=0
    ARG_TYPE=2
    RECURRING=false
    EXECUTION_SCRIPT_LANGUAGE="go"
    EXECUTION_SCRIPT_URL=$IPFS_URL
    echo "Creating Condition-based Dynamic Args Job..."
    ;;
  7)
    # Custom
    TIME_FRAME=36
    TIME_INTERVAL=30
    EXECUTION_SCRIPT_LANGUAGE="ts"
    EXECUTION_SCRIPT_URL=$CUSTOM_IPFS_URL
    echo "Creating Time-base Agentic Job..."
    ;;
  8)
    # Agent Based, Static Args
    TIME_FRAME=36
    TIME_INTERVAL=30
    RECURRING=false
    EXECUTION_SCRIPT_LANGUAGE="ts"
    EXECUTION_SCRIPT_URL=$CUSTOM_IPFS_URL
    echo "Creating Event-based Agentic Job..."
    ;;
  9)
    # Agent Based, Dynamic Args
    TIME_FRAME=36
    TIME_INTERVAL=30
    RECURRING=false
    EXECUTION_SCRIPT_LANGUAGE="ts"
    EXECUTION_SCRIPT_URL=$CUSTOM_IPFS_URL
    echo "Creating Condition-based Agentic Job..."
    ;;
esac

# Set ABI based on target function (default to TEST_FUNCTION_ABI)
# For custom jobs (task definition 7), ABI can be empty
if [ "$TASK_DEFINITION_ID" = "7" ] || [ "$TASK_DEFINITION_ID" = "8" ] || [ "$TASK_DEFINITION_ID" = "9" ]; then
  if [ -z "$ABI_JSON" ]; then
    ABI_JSON='[]'
  fi
  # Ensure ABI_JSON is valid JSON
  ABI_JSON=$(echo "$ABI_JSON" | jq -c .)
else
  if [ -z "$ABI_JSON" ]; then
    ABI_JSON=$(echo "$TEST_FUNCTION_ABI" | jq -c .)
  else
    # Ensure ABI_JSON is valid JSON
    ABI_JSON=$(echo "$ABI_JSON" | jq -c .)
  fi
fi

if [ "$RECURRING" = "true" ]; then
  RECURRING_JSON=true
else
  RECURRING_JSON=false
fi

if [ "$IS_SAFE" = "true" ]; then
  IS_SAFE_JSON=true
else
  IS_SAFE_JSON=false
fi

if ! command -v cast >/dev/null 2>&1; then
  echo "Error: foundry 'cast' CLI is required"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: 'jq' CLI is required for JSON construction"
  exit 1
fi

if [ -n "$EXECUTION_SCRIPT_URL" ]; then
  IPFS_HASH_BYTES32=$(cast keccak "$EXECUTION_SCRIPT_URL")
  if [ $? -ne 0 ] || [ -z "$IPFS_HASH_BYTES32" ]; then
    echo "Error: failed to hash execution script URL"
    exit 1
  fi
else
  EXECUTION_SCRIPT_HASH=""
fi

case $TASK_DEFINITION_ID in
  1)
    JOB_DATA=$(cast abi-encode "encode(uint256)" $TIME_INTERVAL)
    ;;
  2|7)
    if [ -z "$IPFS_HASH_BYTES32" ]; then
      echo "Error: dynamic arguments script URL required for job type $TASK_DEFINITION_ID"
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

echo "Calling createJob() to TriggerXJobRegistry..."

EVENT_SIGNATURE=0x737fc62fbb05dd9fb7c799e68796a2d7c8324e310af7b656c0580e7b3cf8bf8a

echo "Submitting transaction..."

CAST_RESULT=$(cast send \
  --chain $CHAIN_ID \
  --rpc-url $RPC_URL \
  --private-key $SCRIPT_PRIVATE_KEY \
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

# Generate trace ID (UUID format if uuidgen is available, otherwise use random hex)
if command -v uuidgen >/dev/null 2>&1; then
  TRACE_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
else
  # Generate a random 32-character hex string as fallback
  TRACE_ID=$(openssl rand -hex 16)
fi

# Parse ARGUMENTS string into JSON array
if [ -n "$ARGUMENTS" ] && [ "$ARGUMENTS" != "[]" ]; then
  # Validate that ARGUMENTS is valid JSON
  if echo "$ARGUMENTS" | jq . >/dev/null 2>&1; then
    ARGUMENTS_JSON=$(echo "$ARGUMENTS" | jq -c .)
  else
    echo "Warning: ARGUMENTS is not valid JSON, using empty array"
    ARGUMENTS_JSON="[]"
  fi
else
  ARGUMENTS_JSON="[]"
fi

# Set email_id based on recurring and task definition
EMAIL_ID=""
if [ "$RECURRING" = "true" ]; then
  case $TASK_DEFINITION_ID in
    3|4|5|6|8|9)
      # Email required for recurring event/condition jobs
      EMAIL_ID="test@example.com"  # Default, can be overridden via env var
      if [ -n "$USER_EMAIL" ]; then
        EMAIL_ID="$USER_EMAIL"
      fi
      ;;
  esac
fi

# Build base JSON payload (matching CreateJobData struct)
JOB_PAYLOAD=$(jq -n \
  --arg job_id "$JOB_ID" \
  --arg user_address "$USER_ADDRESS" \
  --arg email_id "$EMAIL_ID" \
  --arg job_title "$JOB_TITLE" \
  --arg job_type "$JOB_TYPE" \
  --argjson task_definition_id $TASK_DEFINITION_ID \
  --argjson time_frame $TIME_FRAME \
  --argjson recurring $RECURRING_JSON \
  --arg job_cost_prediction "$JOB_COST_PREDICTION" \
  --arg timezone "$TIMEZONE" \
  --arg created_chain_id "$CREATED_CHAIN_ID" \
  --argjson is_safe $IS_SAFE_JSON \
  --arg safe_address "$SAFE_ADDRESS" \
  --arg safe_name "$SAFE_NAME" \
  --arg schedule_type "$SCHEDULE_TYPE" \
  --argjson time_interval $TIME_INTERVAL \
  --arg cron_expression "$CRON_EXPRESSION" \
  --arg specific_schedule "$SPECIFIC_SCHEDULE" \
  --arg trigger_chain_id "$TRIGGER_CHAIN_ID" \
  --arg trigger_contract_address "$TRIGGER_CONTRACT_ADDRESS" \
  --arg trigger_event "$TRIGGER_EVENT" \
  --arg event_filter_para_name "$EVENT_FILTER_PARA_NAME" \
  --arg event_filter_value "$EVENT_FILTER_VALUE" \
  --arg condition_type "$CONDITION_TYPE" \
  --argjson upper_limit $UPPER_LIMIT \
  --argjson lower_limit $LOWER_LIMIT \
  --arg value_source_type "$VALUE_SOURCE_TYPE" \
  --arg value_source_url "$VALUE_SOURCE_URL" \
  --arg selected_key_route "$SELECTED_KEY_ROUTE" \
  --arg target_chain_id "$TARGET_CHAIN_ID" \
  --arg target_contract_address "$TARGET_CONTRACT_ADDRESS" \
  --arg target_function "$TARGET_FUNCTION" \
  --arg abi "$ABI_JSON" \
  --argjson arg_type $ARG_TYPE \
  --arg arguments_json "$ARGUMENTS_JSON" \
  --arg execution_script_url "$EXECUTION_SCRIPT_URL" \
  --arg execution_script_language "$EXECUTION_SCRIPT_LANGUAGE" \
  --arg execution_script_hash "$EXECUTION_SCRIPT_HASH" \
  --argjson max_execution_time $MAX_EXECUTION_TIME \
  --argjson challenge_period $CHALLENGE_PERIOD \
  '{
    job_id: $job_id,
    user_address: $user_address,
    email_id: $email_id,
    job_title: $job_title,
    job_type: $job_type,
    task_definition_id: $task_definition_id,
    time_frame: $time_frame,
    recurring: $recurring,
    job_cost_prediction: $job_cost_prediction,
    timezone: $timezone,
    created_chain_id: $created_chain_id,
    is_safe: $is_safe,
    arg_type: $arg_type,
    arguments: ($arguments_json | fromjson)
  } + 
  (if ($safe_address | length) > 0 then {safe_address: $safe_address} else {} end) +
  (if ($safe_name | length) > 0 then {safe_name: $safe_name} else {} end) +
  (if ($schedule_type | length) > 0 then {schedule_type: $schedule_type} else {} end) +
  (if $time_interval > 0 then {time_interval: $time_interval} else {} end) +
  (if ($cron_expression | length) > 0 then {cron_expression: $cron_expression} else {} end) +
  (if ($specific_schedule | length) > 0 then {specific_schedule: $specific_schedule} else {} end) +
  (if ($trigger_chain_id | length) > 0 then {trigger_chain_id: $trigger_chain_id} else {} end) +
  (if ($trigger_contract_address | length) > 0 then {trigger_contract_address: $trigger_contract_address} else {} end) +
  (if ($trigger_event | length) > 0 then {trigger_event: $trigger_event} else {} end) +
  (if ($event_filter_para_name | length) > 0 then {event_filter_para_name: $event_filter_para_name} else {} end) +
  (if ($event_filter_value | length) > 0 then {event_filter_value: $event_filter_value} else {} end) +
  (if ($condition_type | length) > 0 then {condition_type: $condition_type} else {} end) +
  (if $upper_limit != 0 then {upper_limit: $upper_limit} else {} end) +
  (if $lower_limit != 0 then {lower_limit: $lower_limit} else {} end) +
  (if ($value_source_type | length) > 0 then {value_source_type: $value_source_type} else {} end) +
  (if ($value_source_url | length) > 0 then {value_source_url: $value_source_url} else {} end) +
  (if ($selected_key_route | length) > 0 then {selected_key_route: $selected_key_route} else {} end) +
  (if ($target_chain_id | length) > 0 then {target_chain_id: $target_chain_id} else {} end) +
  (if ($target_contract_address | length) > 0 then {target_contract_address: $target_contract_address} else {} end) +
  (if ($target_function | length) > 0 then {target_function: $target_function} else {} end) +
  (if ($abi | length) > 0 and $abi != "[]" then {abi: $abi} else {} end) +
  (if ($execution_script_url | length) > 0 then {execution_script_url: $execution_script_url} else {} end) +
  (if ($execution_script_language | length) > 0 then {execution_script_language: $execution_script_language} else {} end) +
  (if ($execution_script_hash | length) > 0 then {execution_script_hash: $execution_script_hash} else {} end) +
  (if $max_execution_time > 0 then {max_execution_time: $max_execution_time} else {} end) +
  (if $challenge_period > 0 then {challenge_period: $challenge_period} else {} end)')

echo "Calling CreateJob API on $DB_SERVER_URL..."
if [ "$SHOW_PAYLOAD" = true ]; then
  echo "Payload:"
  echo "$JOB_PAYLOAD" | jq .
fi

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $DB_SERVER_URL/api/jobs \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: $TRIGGERX_API_KEY" \
  -H "X-Trace-ID: $TRACE_ID" \
  -d "[$JOB_PAYLOAD]")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "HTTP Status: $HTTP_CODE"
if [ "$SHOW_PAYLOAD" = true ]; then
  echo "Response:"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
fi

if [ "$HTTP_CODE" -ge 200 ] && [ "$HTTP_CODE" -lt 300 ]; then
  echo "✓ Job successfully created on server!"
else
  echo "✗ Failed to create job on server"
  exit 1
fi

if [ $TASK_DEFINITION_ID -eq 3 ] || [ $TASK_DEFINITION_ID -eq 4 ]; then
  sleep 10 
  echo "Calling increment() to trigger the event..."

  cast send \
    --chain $CHAIN_ID \
    --rpc-url https://arb-sepolia.g.alchemy.com/v2/$ALCHEMY_API_KEY \
    --private-key $SCRIPT_PRIVATE_KEY \
    $TEST_CONTRACT_ADDRESS "increment()"
fi
