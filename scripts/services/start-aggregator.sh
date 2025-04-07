#!/bin/bash

source .env

# Delay in ms for the aggregator to wait before submitting the task to the EigenLayer
DELAY=10000

# Sync interval in ms for the aggregator to run internal tasks
SYNC_INTERVAL=7200000

othentic-cli node aggregator \
    --json-rpc \
    --json-rpc.port $AGGREGATOR_RPC_PORT \
    --json-rpc.custom-message-enabled \
    --p2p.port $AGGREGATOR_P2P_PORT \
    --p2p.datadir data/peerstore/aggregator \
    --p2p.discovery-interval 10000 \
    --internal-tasks \
    --sync-interval 60000 \
    --metrics \
    --delay $DELAY \
    --keystore .keystore/aggregator.json 