#!/bin/bash

curl -X POST http://localhost:8080/api/quorums \
  -H "Content-Type: application/json" \
  -d "{
    \"quorum_id\": 5,
    \"quorum_no\": 4,
    \"quorum_creation_block\": 1,
    \"quorum_tx_hash\": \"\",
    \"keepers\": [\"\"],
    \"quorum_stake_total\": 0,
    \"task_ids\": [],
    \"quorum_status\": true
}"