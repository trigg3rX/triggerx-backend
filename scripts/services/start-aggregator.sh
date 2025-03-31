#!/bin/bash

source .env

# Delay in ms for the aggregator to wait before submitting the task to the EigenLayer
DELAY=10000

# Sync interval in ms for the aggregator to run internal tasks
SYNC_INTERVAL=7200000

othentic-cli node aggregator \
    --json-rpc \
    --json-rpc.port $AGGREGATOR_RPC_PORT \
    --p2p.port $AGGREGATOR_P2P_PORT \
    --p2p.datadir data/peerstore \
    --internal-tasks \
    --metrics \
    --delay $DELAY \
    --sync-interval $SYNC_INTERVAL \
    --keystore .keystore/aggregator.json \
    --json-rpc.custom-message-enabled