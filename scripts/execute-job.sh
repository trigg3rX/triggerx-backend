#! /bin/bash


curl -X POST http://localhost:4003/task/execute \
  -H "Content-Type: application/json" \
  -d "{
    \"taskDefinitionId\": 1,
    \"job\": {
      \"job_id\": 999,
      \"targetFunction\": \"execute\",
      \"arguments\": \"bleh\",
      \"chainID\": 11155420,
      \"contractAddress\": \"0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d\"
    }
  }"