#!/bin/bash

curl -X POST http://localhost:9003/scheduler/submit-task \
  -H "Content-Type: application/json" \
  -d "{
        \"source\": \"time_scheduler\",
        \"scheduler_id\": 1234,
        \"send_task_data_to_keeper\": {
            \"task_id\": 1,
            \"performer_data\": {
                \"keeper_id\": 3,
                \"keeper_address\": \"0x7Db951c0E6D8906687B459427eA3F3F2b456473B\"
            },
            \"target_data\": [
                {
                    \"job_id\": 1,
                    \"task_id\": 1,
                    \"task_definition_id\": 1,
                    \"target_chain_id\": \"11155420\",
                    \"target_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
                    \"target_function\": \"incrementBy\",
                    \"abi\": \"[{\\\"anonymous\\\":false,\\\"inputs\\\":[{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"previousValue\\\",\\\"type\\\":\\\"uint256\\\"},{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"newValue\\\",\\\"type\\\":\\\"uint256\\\"},{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"incrementAmount\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"name\\\":\\\"CounterIncremented\\\",\\\"type\\\":\\\"event\\\"},{\\\"inputs\\\":[],\\\"name\\\":\\\"getCount\\\",\\\"outputs\\\":[{\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"stateMutability\\\":\\\"view\\\",\\\"type\\\":\\\"function\\\"},{\\\"inputs\\\":[],\\\"name\\\":\\\"increment\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"},{\\\"inputs\\\":[{\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"amount\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"name\\\":\\\"incrementBy\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
                    \"arg_type\": 1,
                    \"is_imua\": true,
                    \"arguments\": [
                        \"3\"
                    ],
                    \"dynamic_arguments_script_url\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreia2figqi2fme3gw5qauzgmpkky4exmvmvpox6qlf45rjsmd6qqpum\"
                }
            ],
            \"trigger_data\": [
                {
                    \"task_id\": 1,
                    \"task_definition_id\": 1,
                    \"recurring\": false,
                    \"expiration_time\": \"2025-07-22T23:59:59Z\",
                    \"trigger_timestamp\": \"2025-07-22T15:00:00Z\",
                    \"next_trigger_timestamp\": \"2025-07-22T16:00:00Z\",
                    \"time_interval\": 3600
                }
            ],
            \"scheduler_signature_data\": {
                \"task_id\": 1,
                \"scheduler_id\": 1234,
                \"scheduler_signing_address\": \"0x7Db951c0E6D8906687B459427eA3F3F2b456473B\",
                \"scheduler_signature\": \"0x1234567890\"
            }
        }
    }"
