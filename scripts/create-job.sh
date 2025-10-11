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
    TIME_FRAME=2500
    TIME_INTERVAL=0
    ARG_TYPE=1
    DYNAMIC_ARGUMENTS_SCRIPT_URL=
    echo "Creating Event-based Static Args Job..."
    ;;
  4)
    TASK_DEFINITION_ID=4 # Event Based, Dynamic Args
    TIME_FRAME=2500
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


echo "\nCalling CreateJob() to TriggerXJobRegistry..."

cast send \
  --async \
  --chain 84532 \
  --rpc-url https://base-sepolia.g.alchemy.com/v2/PIUHuKF0BQoK9ibzzDH-B-bk0LbiYY38 \
  --private-key 2212ec53a0dddee799b3342a86bb45fd1192a1981139244869102be2b3c47045 \
  0xdB66c11221234C6B19cCBd29868310c31494C21C \
  "createJob(string,uint8,uint256,address,bytes)" test $TASK_DEFINITION_ID $TIME_FRAME 0x49a81A591afdDEF973e6e49aaEa7d76943ef234C 0x01 \
  -- --broadcast

sleep 3

# curl -X POST https://data.triggerx.network/api/jobs \ 192.168.1.56
curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: ADMIN" \
  -d "[
    {
      \"user_address\": \"0x7Db951c0E6D8906687B459427eA3F3F2b456473B\",
      \"ether_balance\": 50000000000000000,
      \"token_balance\": 50000000000000000000,
      \"created_chain_id\": \"421614\",
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
      \"trigger_contract_address\": \"0x980B62Da83eFf3D4576C647993b0c1D7faf17c73\",
      \"trigger_event\": \"Transfer(address,address,uint256)\",
      \"event_filter_para_name\": \"to\",
      \"event_filter_value\": \"0xC9dC9c361c248fFA0890d7E1a263247670914980\",
      \"condition_type\": \"less_than\",
      \"upper_limit\": 92,
      \"lower_limit\": 89,
      \"value_source_type\": \"api\",
      \"value_source_url\": \"http://localhost:8080/price\",
      \"target_chain_id\": \"421614\",
      \"target_contract_address\": \"0x980B62Da83eFf3D4576C647993b0c1D7faf17c73\",
      \"target_function\": \"implementation\",
      \"abi\": \"[{\\\"inputs\\\":[],\\\"name\\\":\\\"implementation\\\",\\\"outputs\\\":[{\\\"internalType\\\":\\\"address\\\",\\\"name\\\":\\\"implementation_\\\",\\\"type\\\":\\\"address\\\"}],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
      \"arg_type\": $ARG_TYPE,
      \"is_imua\": false,
      \"arguments\": [\"3\"],
      \"created_chain_id\": \"421614\",
      \"dynamic_arguments_script_url\": \"$DYNAMIC_ARGUMENTS_SCRIPT_URL\"
    }
  ]"

echo "\n"

if [ $TASK_DEFINITION_ID -eq 3 ] || [ $TASK_DEFINITION_ID -eq 4 ]; then
  sleep 5 
  echo "\nCalling increment() to trigger the event..."

  cast send \
    --async \
    --chain 421614 \
    --rpc-url https://arb-sepolia.g.alchemy.com/v2/PIUHuKF0BQoK9ibzzDH-B-bk0LbiYY38 \
    --private-key 2212ec53a0dddee799b3342a86bb45fd1192a1981139244869102be2b3c47045 \
    0x980B62Da83eFf3D4576C647993b0c1D7faf17c73 "increment()" \
    -- --broadcast
fi

echo "\n"