#!/bin/bash

# TIME BASED JOB
echo "Creating Time-based Job..."
curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"user_address\": \"0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9D\",
      \"stake_amount\": 1000000000000000000,
      \"token_amount\": 1000000000000000000,
      \"task_definition_id\": 1,
      \"priority\": 5,
      \"security\": 5,
      \"custom\": false,
      \"job_title\": \"Test Time Job\",
      \"time_frame\": 3600,
      \"recurring\": true,
      \"time_interval\": 1800,
      \"job_cost_prediction\": 0.1,
      \"target_chain_id\": \"1\",
      \"target_contract_address\": \"0x1234567890123456789012345678901234567890\",
      \"target_function\": \"execute\",
      \"abi\": \"[{\\\"inputs\\\":[],\\\"name\\\":\\\"execute\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
      \"arg_type\": 1,
      \"arguments\": [],
      \"job_type\": \"time\",
      \"timezone\": \"UTC\"
    }
  ]"

echo -e "\nCreating Event-based Job..."
# EVENT BASED JOB
curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"user_address\": \"0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9F\",
      \"stake_amount\": 1000000000000000000,
      \"token_amount\": 1000000000000000000,
      \"task_definition_id\": 3,
      \"priority\": 5,
      \"security\": 5,
      \"custom\": false,
      \"job_title\": \"Test Event Job\",
      \"time_frame\": 3600,
      \"recurring\": false,
      \"trigger_chain_id\": \"1\",
      \"trigger_contract_address\": \"0x1234567890123456789012345678901234567890\",
      \"trigger_event\": \"PriceUpdated(uint256 price, uint256 timestamp)\",
      \"target_chain_id\": \"1\",
      \"target_contract_address\": \"0x1234567890123456789012345678901234567890\",
      \"target_function\": \"updatePrice\",
      \"abi\": \"[{\\\"inputs\\\":[{\\\"name\\\":\\\"price\\\",\\\"type\\\":\\\"uint256\\\"},{\\\"name\\\":\\\"timestamp\\\",\\\"type\\\":\\\"uint256\\\"}],\\\"name\\\":\\\"PriceUpdated\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
      \"arg_type\": 1,
      \"arguments\": [\"1000000000000000000\"],
      \"job_type\": \"event\",
      \"timezone\": \"UTC\",
      \"job_cost_prediction\": 0.1
    }
  ]"

echo -e "\nCreating Condition-based Job..."
# CONDITION BASED JOB
curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"user_address\": \"0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9E\",
      \"stake_amount\": 1000000000000000000,
      \"token_amount\": 1000000000000000000,
      \"task_definition_id\": 5,
      \"priority\": 5,
      \"security\": 5,
      \"custom\": false,
      \"job_title\": \"Test Condition Job\",
      \"time_frame\": 3600,
      \"recurring\": false,
      \"condition_type\": \"price\",
      \"upper_limit\": 1000000000000000000,
      \"lower_limit\": 500000000000000000,
      \"value_source_type\": \"api\",
      \"value_source_url\": \"https://api.example.com/price\",
      \"target_chain_id\": \"1\",
      \"target_contract_address\": \"0x1234567890123456789012345678901234567890\",
      \"target_function\": \"execute\",
      \"abi\": \"[{\\\"inputs\\\":[],\\\"name\\\":\\\"execute\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
      \"arg_type\": 1,
      \"arguments\": [],
      \"job_type\": \"condition\",
      \"timezone\": \"UTC\",
      \"job_cost_prediction\": 0.1
    }
  ]"