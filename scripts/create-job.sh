#!/bin/bash

# TIME BASED JOB
curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"user_address\": \"0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9E\",
      \"stake_amount\": 101,
      \"token_amount\": 111,
      \"task_definition_id\": 1,
      \"priority\": 1,
      \"security\": 1,
      \"time_frame\": 259200,
      \"recurring\": true,
      \"time_interval\": 120,
      \"trigger_chain_id\": \"11155420\",
      \"trigger_contract_address\": \"0xf9f40AA5436304EC1d9e84fc85256aE80086E3a1\",
      \"trigger_event\": \"PriceUpdated(uint256 price, uint256 timestamp)\",
      \"script_ipfs_url\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreicimunflmfgplxovjaghko5moubvzacoedhu3bqbcs37ox2ypzgbe\",
      \"script_trigger_function\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreihbtjghsjl3cpavwyvdqazeto2y5blmg5szdi3djozgbr5odoso5a\",
      \"target_chain_id\": \"11155420\",
      \"target_contract_address\": \"0xf9f40AA5436304EC1d9e84fc85256aE80086E3a1\",
      \"target_function\": \"updatePrice\",
      \"arg_type\": 1,
      \"arguments\": [\"19\"],
      \"script_target_function\": \"checker\",
      \"job_cost_prediction\": 253
    }
  ]"

# EVENT BASED JOB


# CONDITION BASED JOB, NEVER SATISFIED

# curl -X POST http://192.168.1.17:9002/api/jobs \
#   -H "Content-Type: application/json" \
#   -d "[
#     {
#       \"user_address\": \"0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9E\",
#       \"stake_amount\": 101,
#       \"token_amount\": 111,
#       \"task_definition_id\": 5,
#       \"priority\": 1,
#       \"security\": 1,
#       \"time_frame\": 70,
#       \"recurring\": true,
#       \"time_interval\": 30,
#       \"trigger_chain_id\": \"11155420\",
#       \"trigger_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
#       \"trigger_event\": \"CounterIncremented(uint256,uint256,uint256)\",
#       \"script_ipfs_url\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreiekuwtrifkzqswkyjj7urie23vmv2fl43p3dzbu53yye3xmdam7p4\",
#       \"script_trigger_function\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreihbtjghsjl3cpavwyvdqazeto2y5blmg5szdi3djozgbr5odoso5a\",
#       \"target_chain_id\": \"11155420\",
#       \"target_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
#       \"target_function\": \"increment\",
#       \"arg_type\": 1,
#       \"arguments\": [\"19\"],
#       \"script_target_function\": \"checker\",
#       \"job_cost_prediction\": 253
#     }
#   ]"

# CONDITION BASED JOB, SATISFIED, AA ENABLED

# 192.168.1.17
curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"user_address\": \"0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9E\",
      \"stake_amount\": 101,
      \"token_amount\": 111,
      \"task_definition_id\": 5,
      \"priority\": 1,
      \"security\": 1,
      \"time_frame\": 300,
      \"recurring\": false,
      \"time_interval\": 60,
      \"trigger_chain_id\": \"11155420\",
      \"trigger_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
      \"trigger_event\": \"CounterIncremented(uint256,uint256,uint256)\",
      \"script_ipfs_url\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreicimunflmfgplxovjaghko5moubvzacoedhu3bqbcs37ox2ypzgbe\",
      \"script_trigger_function\": \"https://teal-random-koala-993.mypinata.cloud/ipfs/bafkreif426p7t7takzhw3g6we2h6wsvf27p5jxj3gaiynqf22p3jvhx4la\",
      \"target_chain_id\": \"11155420\",
      \"target_contract_address\": \"0x49a81A591afdDEF973e6e49aaEa7d76943ef234C\",
      \"target_function\": \"increment\",
      \"abi\": \"[{\\\"inputs\\\":[],\\\"name\\\":\\\"increment\\\",\\\"outputs\\\":[],\\\"stateMutability\\\":\\\"nonpayable\\\",\\\"type\\\":\\\"function\\\"}]\",
      \"arg_type\": 1,
      \"arguments\": [\"19\",\"5\"],
      \"script_target_function\": \"checker\",
      \"job_cost_prediction\": 253
    }
  ]"