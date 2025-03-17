#!/bin/bash

# Array of user addresses
USER_ADDRESSES=(
    "0xE3304AB782c3272D1D7964ba7043e11Fd7ecFEB8"
    "0x2F5e2E9C62F2A0b9C95D56C43e8A3B075f5A4e1D"
    "0xB1c4D2f8E69A1c5d4a52C8bB962b9E4b2F8D5e3A"
    "0x6D9f7A4E3B2C1a8F5e0D6B9c4A3E8d2F1B5c7D9E"
)

# Get random user address
RANDOM_INDEX=$((RANDOM % ${#USER_ADDRESSES[@]}))
SELECTED_USER_ADDRESS=${USER_ADDRESSES[$RANDOM_INDEX]}

GAS_PRICE=4

CHAIN_ID=11155420

JOB_TYPE=(1 2 3)
RANDOM_JOB_TYPE_INDEX=$((RANDOM % ${#JOB_TYPE[@]}))
SELECTED_JOB_TYPE=${JOB_TYPE[$RANDOM_JOB_TYPE_INDEX]}

# Array of job cost predictions
JOB_COST_PREDICTIONS=(208 219 256 303)

# Get random job cost prediction
RANDOM_COST_INDEX=$((RANDOM % ${#JOB_COST_PREDICTIONS[@]}))
SELECTED_JOB_COST=${JOB_COST_PREDICTIONS[$RANDOM_COST_INDEX]}

# Create 3 new jobs
# curl -X POST http://localhost:8080/api/jobs \
#   -H "Content-Type: application/json" \
#   -d "[
#     {
#       \"user_address\": \"$SELECTED_USER_ADDRESS\",
#       \"stake_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"token_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"task_definition_id\": $((1 + RANDOM % 3)),
#       \"priority\": 1,
#       \"security\": 1,
#       \"time_frame\": 5,
#       \"recurring\": true,
#       \"time_interval\": 10,
#       \"trigger_chain_id\": \"1\",
#       \"trigger_contract_address\": \"0xF1d505d1f6df11795c77A8A1b7476609E7b6361a\",
#       \"trigger_event\": \"Staked(address indexed user, uint256 amount)\",
#       \"script_ipfs_url\": \"https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a\",
#       \"script_trigger_function\": \"checker\",
#       \"target_chain_id\": \"1\",
#       \"target_contract_address\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\",
#       \"target_function\": \"execute\",
#       \"arg_type\": 1,
#       \"arguments\": [\"19\", \"91\"],
#       \"script_target_function\": \"checker\",
#       \"job_cost_prediction\": $SELECTED_JOB_COST
#     },
#     {
#       \"user_address\": \"$SELECTED_USER_ADDRESS\",
#       \"stake_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"token_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"task_definition_id\": $((1 + RANDOM % 3)),
#       \"priority\": 2,
#       \"security\": 1,
#       \"time_frame\": 10,
#       \"recurring\": true,
#       \"time_interval\": 15,
#       \"trigger_chain_id\": \"1\",
#       \"trigger_contract_address\": \"0xF1d505d1f6df11795c77A8A1b7476609E7b6361a\",
#       \"trigger_event\": \"Transfer(address indexed from, address indexed to, uint256 value)\",
#       \"scriptIPFSUrl\": \"https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a\",
#       \"scriptTriggerFunction\": \"checker\",
#       \"targetChainID\": \"1\",
#       \"targetContractAddress\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\",
#       \"targetFunction\": \"transfer\",
#       \"argType\": 1,
#       \"arguments\": [\"100\"],
#       \"scriptTargetFunction\": \"checker\",
#       \"jobCostPrediction\": $SELECTED_JOB_COST
#     },
#     {
#       \"userAddress\": \"$SELECTED_USER_ADDRESS\",
#       \"stakeAmount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"tokenAmount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"taskDefinitionID\": $((1 + RANDOM % 3)),
#       \"priority\": 1,
#       \"security\": 2,
#       \"timeFrame\": 15,
#       \"recurring\": false,
#       \"timeInterval\": 20,
#       \"triggerChainID\": \"1\",
#       \"triggerContractAddress\": \"0xF1d505d1f6df11795c77A8A1b7476609E7b6361a\",
#       \"triggerEvent\": \"Approval(address indexed owner, address indexed spender, uint256 value)\",
#       \"scriptIPFSUrl\": \"https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a\",
#       \"scriptTriggerFunction\": \"checker\",
#       \"targetChainID\": \"1\",
#       \"targetContractAddress\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\",
#       \"targetFunction\": \"approve\",
#       \"argType\": 1,
#       \"arguments\": [\"200\"],
#       \"scriptTargetFunction\": \"checker\",
#       \"jobCostPrediction\": $SELECTED_JOB_COST
#     }
#   ]"

# # Create 2 new jobs
# curl -X POST http://localhost:8080/api/jobs \
#   -H "Content-Type: application/json" \
#   -d "[
#     {
#       \"user_address\": \"$SELECTED_USER_ADDRESS\",
#       \"stake_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"token_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"task_definition_id\": $((1 + RANDOM % 3)),
#       \"priority\": 1,
#       \"security\": 1,
#       \"time_frame\": 5,
#       \"recurring\": true,
#       \"time_interval\": 10,
#       \"trigger_chain_id\": \"1\",
#       \"trigger_contract_address\": \"0xF1d505d1f6df11795c77A8A1b7476609E7b6361a\",
#       \"trigger_event\": \"Staked(address indexed user, uint256 amount)\",
#       \"script_ipfs_url\": \"https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a\",
#       \"script_trigger_function\": \"checker\",
#       \"target_chain_id\": \"1\",
#       \"target_contract_address\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\",
#       \"target_function\": \"execute\",
#       \"arg_type\": 1,
#       \"arguments\": [\"19\", \"91\"],
#       \"script_target_function\": \"checker\",
#       \"job_cost_prediction\": $SELECTED_JOB_COST
#     },
#     {
#       \"user_address\": \"$SELECTED_USER_ADDRESS\",
#       \"stake_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"token_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
#       \"task_definition_id\": $((1 + RANDOM % 3)),
#       \"priority\": 2,
#       \"security\": 1,
#       \"time_frame\": 10,
#       \"recurring\": true,
#       \"time_interval\": 15,
#       \"trigger_chain_id\": \"1\",
#       \"trigger_contract_address\": \"0xF1d505d1f6df11795c77A8A1b7476609E7b6361a\",
#       \"trigger_event\": \"Transfer(address indexed from, address indexed to, uint256 value)\",
#       \"scriptIPFSUrl\": \"https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a\",
#       \"scriptTriggerFunction\": \"checker\",
#       \"targetChainID\": \"1\",
#       \"targetContractAddress\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\",
#       \"targetFunction\": \"transfer\",
#       \"argType\": 1,
#       \"arguments\": [\"100\"],
#       \"scriptTargetFunction\": \"checker\",
#       \"jobCostPrediction\": $SELECTED_JOB_COST
#     }
#   ]"

# Create 1 new job
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d "[
    {
      \"user_address\": \"$SELECTED_USER_ADDRESS\",
      \"stake_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
      \"token_amount\": $((GAS_PRICE * SELECTED_JOB_COST)),
      \"task_definition_id\": 6,
      \"priority\": 1,
      \"security\": 1,
      \"time_frame\": 60,
      \"recurring\": true,
      \"time_interval\": 3,
      \"trigger_chain_id\": \"1\",
      \"trigger_contract_address\": \"0xf9f40AA5436304EC1d9e84fc85256aE80086E3a1\",
      \"trigger_event\": \"Staked(address indexed user, uint256 amount)\",
      \"script_ipfs_url\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreiekuwtrifkzqswkyjj7urie23vmv2fl43p3dzbu53yye3xmdam7p4\",
      \"script_trigger_function\": \"https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/bafkreihbtjghsjl3cpavwyvdqazeto2y5blmg5szdi3djozgbr5odoso5a\",
      \"target_chain_id\": \"1\",
      \"target_contract_address\": \"0xf9f40AA5436304EC1d9e84fc85256aE80086E3a1\",
      \"target_function\": \"updatePrice\",
      \"arg_type\": 1,
      \"arguments\": [\"19\"],
      \"script_target_function\": \"checker\",
      \"job_cost_prediction\": $SELECTED_JOB_COST
    }
  ]"
