#!/bin/bash

# Array of keeper addresses
KEEPER_ADDRESSES=(
    # "0x7789fEa5E9a6472cA38A0ADc62EF7B4e270Fe89D"
    # "0x350ddC16818CDFA1E0c06c26Fd5e76360032faa4"
    # "0x00ACe36D365e9Ed147538Cf10C47BfAe1c9B2271"
    "0x011FCbAE5f306cd793456ab7d4c0CC86756c693D"
)

REGISTERED_TXS="0x1234567890abcdef"

CONSENSUS_KEYS="3bd66a68dcde6ede3b38ced6de79489a447e0fac1648b749a5001b0aa167d089"

CONNECTION_ADDRESS="http://localhost:9005/task/execute"

for KEEPER_ADDRESS in "${KEEPER_ADDRESSES[@]}"; do
    echo "Creating keeper with address: $KEEPER_ADDRESS"

    # Create keeper
    curl -X POST http://localhost:8080/api/keepers \
        -H "Content-Type: application/json" \
        -d "{
            \"keeper_address\": \"$KEEPER_ADDRESS\",
            \"registered_tx\": \"$REGISTERED_TX\",
            \"rewards_address\": \"$KEEPER_ADDRESS\",
            \"consensus_keys\": [\"$CONSENSUS_KEYS\"],
            \"connection_address\": \"$CONNECTION_ADDRESS\"
        }"
done
