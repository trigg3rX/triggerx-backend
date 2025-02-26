#!/bin/bash

# Array of keeper addresses
KEEPER_ADDRESSES=(
    # "0x7789fEa5E9a6472cA38A0ADc62EF7B4e270Fe89D"
    # "0x350ddC16818CDFA1E0c06c26Fd5e76360032faa4"
    # "0x00ACe36D365e9Ed147538Cf10C47BfAe1c9B2271"
    "0xFf22567685caF4De37eFf02159dB5D2B39A65C82"
)

REGISTERED_TXS="0x1234567890abcdef"

CONSENSUS_KEYS="0x1234567890abcdef"

for KEEPER_ADDRESS in "${KEEPER_ADDRESSES[@]}"; do
    echo "Creating keeper with address: $KEEPER_ADDRESS"

    # Create keeper
    curl -X POST http://localhost:9004/api/keepers \
        -H "Content-Type: application/json" \
        -d "{
            \"keeper_address\": \"$KEEPER_ADDRESS\",
            \"registered_tx\": \"$REGISTERED_TX\",
            \"rewards_address\": \"$KEEPER_ADDRESS\",
            \"consensus_keys\": [\"$CONSENSUS_KEYS\"]
        }"
done
