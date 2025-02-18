#! /bin/bash

JOB_TYPE=(1 2 3 4 5 6)
RANDOM_JOB_TYPE_INDEX=$((RANDOM % ${#JOB_TYPE[@]}))
SELECTED_JOB_TYPE=${JOB_TYPE[$RANDOM_JOB_TYPE_INDEX]}

curl -X POST http://localhost:4003/task/execute \
  -H "Content-Type: application/json" \
  -d "{
    \"job_id\": 999,
    \"job_type\": 1,
    \"chain_id\": 11155420,
    \"time_interval\": 300,
    \"trigger_contract_address\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\",
    \"trigger_event\": \"TaskExecuted\",
    \"target_contract_address\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\", 
    \"target_function\": \"execute\",
    \"arg_type\": 1,
    \"arguments\": [\"arg1\", \"arg2\"],
    \"script_function\": \"\",
    \"script_ipfs_url\": \"\",
    \"task_id\": 1,
    \"task_definition_id\": $SELECTED_JOB_TYPE,
    \"task_performer\": \"0x1234567890123456789012345678901234567890\"
  }"