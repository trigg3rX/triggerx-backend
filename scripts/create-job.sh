#!/bin/bash

# for different combinations:
# task_definition_id: 1, 2, 3, 4, 5, 6
# recurring: true, false
# schedule_type: interval, cron, specific
# condition_type: price, volume
# arg_type: 0, 1, 2

echo "Creating Time-based Job..."
# curl -X POST http://192.168.1.56:9002/api/jobs \
curl -X POST https://data.triggerx.network/api/jobs \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"user_address\": \"0x7Db951c0E6D8906687B459427eA3F3F2b456473B\",
      \"ether_balance\": 50000000000000000,
      \"token_balance\": 50000000000000000000,
      \"job_title\": \"Wind's Howling\",
      \"task_definition_id\": 1,
      \"custom\": true,
      \"time_frame\": 35,
      \"recurring\": false,
      \"job_cost_prediction\": 0.1,
      \"timezone\": \"IST\",
      \"schedule_type\": \"interval\",
      \"time_interval\": 30,
      \"cron_expression\": \"0 0 * * *\",
      \"specific_schedule\": \"2025-01-01 00:00:00\",
      \"trigger_chain_id\": \"11155420\",
      \"trigger_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
      \"trigger_event\": \"CounterIncremented(uint256 previousValue,uint256 newValue,uint256 incrementAmount)\",
      \"condition_type\": \"price\",
      \"upper_limit\": 2600,
      \"lower_limit\": 2400,
      \"value_source_type\": \"api\",
      \"value_source_url\": \"https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd\",
      \"target_chain_id\": \"11155420\",
      \"target_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
      \"target_function\": \"incrementBy\",
      \"abi\": \"[{\\\"anonymous\\\":false,\\\"inputs\\\":[{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"previousValue\\\",\\\"type\\\":\\\"uint256\\\"},{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"newValue\\\",\\\"type\\\":\\\"uint256\\\"},{\\\"indexed\\\":false,\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"incrementAmount\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"name\\\":\\\"CounterIncremented\\\",\\\"type\\\":\\\"event\\\"},{\\\"inputs\\\":[],\\\"name\\\":\\\"getCount\\\",\\\"outputs\\\":[{\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"stateMutability\\\":\\\"view\\\",\\\"type\\\":\\\"function\\\"},{\\\"inputs\\\":[],\\\"name\\\":\\\"increment\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"},{\\\"inputs\\\":[{\\\"internalType\\\":\\\"uint256\\\",\\\"name\\\":\\\"amount\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"name\\\":\\\"incrementBy\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
      \"arg_type\": 1,
      \"arguments\": [\"3\"],
      \"dynamic_arguments_script_url\": \"https://teal-random-koala-993.mypinata.cloud/ipfs/bafkreif426p7t7takzhw3g6we2h6wsvf27p5jxj3gaiynqf22p3jvhx4la\"
    }
  ]"
