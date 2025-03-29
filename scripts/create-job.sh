#!/bin/bash

curl -X POST http://localhost:9002/api/jobs \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"user_address\": \"0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9E\",
      \"stake_amount\": 101,
      \"token_amount\": 111,
      \"task_definition_id\": 2,
      \"priority\": 1,
      \"security\": 1,
      \"time_frame\": 2,
      \"recurring\": true,
      \"time_interval\": 2,
      \"trigger_chain_id\": \"11155420\",
      \"trigger_contract_address\": \"0xf9f40AA5436304EC1d9e84fc85256aE80086E3a1\",
      \"trigger_event\": \"PriceUpdated(uint256 price, uint256 timestamp)\",
      \"script_ipfs_url\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreiekuwtrifkzqswkyjj7urie23vmv2fl43p3dzbu53yye3xmdam7p4\",
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
