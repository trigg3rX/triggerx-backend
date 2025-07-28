#!/bin/bash

if [ $# -ne 1 ]; then
  echo "Usage: $0 <job_type (1-6)>"
  exit 1
fi

TASK_DEFINITION_ID=$1

if ! [[ "$TASK_DEFINITION_ID" =~ ^[1-6]$ ]]; then
  echo "Error: job_type must be an integer between 1 and 6."
  exit 1
fi

adj=$(shuf -n 1 scripts/adjectives.txt)
noun=$(shuf -n 1 scripts/nouns.txt)

JOB_TITLE="$adj $noun"

JOB_ID=$(date +%s)

case $TASK_DEFINITION_ID in
  1)
    TASK_DEFINITION_ID=1 # Time Based, Static Args
    TIME_FRAME=35
    TIME_INTERVAL=32
    ARG_TYPE=1
    DYNAMIC_ARGUMENTS_SCRIPT_URL=
    echo "Creating Time-based Static Args Job..."
    ;;
  2)
    TASK_DEFINITION_ID=2 # Time Based, Dynamic Args
    TIME_FRAME=35
    TIME_INTERVAL=32
    ARG_TYPE=2
    DYNAMIC_ARGUMENTS_SCRIPT_URL="https://teal-random-koala-993.mypinata.cloud/ipfs/bafkreif426p7t7takzhw3g6we2h6wsvf27p5jxj3gaiynqf22p3jvhx4la"
    echo "Creating Time-based Dynamic Args Job..."
    ;;
  3)
    TASK_DEFINITION_ID=3 # Event Based, Static Args
    TIME_FRAME=20
    TIME_INTERVAL=0
    ARG_TYPE=1
    DYNAMIC_ARGUMENTS_SCRIPT_URL=
    echo "Creating Event-based Static Args Job..."
    ;;
  4)
    TASK_DEFINITION_ID=4 # Event Based, Dynamic Args
    TIME_FRAME=20
    TIME_INTERVAL=0
    ARG_TYPE=2
    DYNAMIC_ARGUMENTS_SCRIPT_URL="https://teal-random-koala-993.mypinata.cloud/ipfs/bafkreif426p7t7takzhw3g6we2h6wsvf27p5jxj3gaiynqf22p3jvhx4la"
    echo "Creating Event-based Dynamic Args Job..."
    ;;
  5)
    TASK_DEFINITION_ID=5 # Condition Based, Static Args
    TIME_FRAME=20
    TIME_INTERVAL=0
    ARG_TYPE=1
    DYNAMIC_ARGUMENTS_SCRIPT_URL=
    echo "Creating Condition-based Static Args Job..."
    ;;
  6)
    TASK_DEFINITION_ID=6 # Condition Based, Dynamic Args
    TIME_FRAME=20
    TIME_INTERVAL=0
    ARG_TYPE=2
    DYNAMIC_ARGUMENTS_SCRIPT_URL="https://teal-random-koala-993.mypinata.cloud/ipfs/bafkreif426p7t7takzhw3g6we2h6wsvf27p5jxj3gaiynqf22p3jvhx4la"
    echo "Creating Condition-based Dynamic Args Job..."
    ;;
esac

# curl -X POST https://data.triggerx.network/api/jobs \ 192.168.1.56
curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: admin-1234-xyz" \
  -d "[
    {
      \"user_address\": \"0x7Db951c0E6D8906687B459427eA3F3F2b456473B\",
      \"ether_balance\": 50000000000000000,
      \"token_balance\": 50000000000000000000,
      \"created_chain_id\": \"11155420\",
      \"job_id\": $JOB_ID,
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
      \"trigger_chain_id\": \"11155420\",
      \"trigger_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
      \"trigger_event\": \"CounterIncremented(uint256,uint256,uint256)\",
      \"condition_type\": \"less_than\",
      \"upper_limit\": 92,
      \"lower_limit\": 89,
      \"value_source_type\": \"api\",
      \"value_source_url\": \"http://localhost:8080/price\",
      \"target_chain_id\": \"11155420\",
      \"target_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
      \"target_function\": \"incrementBy\",
      \"abi\": \"[{\\\"anonymous\\\":false,\\\"inputs\\\":[{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"previousValue\\\",\\\"type\\\":\\\"uint256\\\"},{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"newValue\\\",\\\"type\\\":\\\"uint256\\\"},{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"incrementAmount\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"name\\\":\\\"CounterIncremented\\\",\\\"type\\\":\\\"event\\\"},{\\\"inputs\\\":[],\\\"name\\\":\\\"getCount\\\",\\\"outputs\\\":[{\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"stateMutability\\\":\\\"view\\\",\\\"type\\\":\\\"function\\\"},{\\\"inputs\\\":[],\\\"name\\\":\\\"increment\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"},{\\\"inputs\\\":[{\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"amount\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"name\\\":\\\"incrementBy\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
      \"arg_type\": $ARG_TYPE,
      \"is_imua\": true,
      \"arguments\": [\"3\"],
      \"dynamic_arguments_script_url\": \"$DYNAMIC_ARGUMENTS_SCRIPT_URL\"
    }
  ]"

if [ $TASK_DEFINITION_ID -eq 3 ] || [ $TASK_DEFINITION_ID -eq 4 ]; then
  sleep 5
  echo "\nCalling increment() to trigger the event..."

  cast send \
    --async \
    --chain 11155420 \
    --rpc-url https://opt-sepolia.g.alchemy.com/v2/PIUHuKF0BQoK9ibzzDH-B-bk0LbiYY38 \
    --private-key 3a636c73c3388970114d86ff4d5f0becffaff0db24db342eebb00323238f0fda \
    0x49a81A591afdDEF973e6e49aaEa7d76943ef234C "increment()" \
    -- --broadcast
fi
