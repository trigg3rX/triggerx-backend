#!/bin/bash

# Delay in ms for the aggregator to wait before submitting the task to the EigenLayer
DELAY=10000

# Sync interval in ms for the aggregator to run internal tasks
SYNC_INTERVAL=7200000

# Metrics port, default 6060
METRICS_PORT=6060

# Metrics export url
METRICS_EXPORT_URL=

othentic-cli node aggregator \
    --json-rpc \
    --json-rpc.port $AGGREGATOR_RPC_PORT \
    --p2p.port $AGGREGATOR_P2P_PORT \
    --internal-tasks \
    --metrics \
    --delay $DELAY \
    --sync-interval $SYNC_INTERVAL \
    --keystore .keystore/aggregator.json \
    --json-rpc.custom-message-enabled