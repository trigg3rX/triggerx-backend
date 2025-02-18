#!/bin/bash


# Delay in ms for the aggregator to wait before submitting the task to the EigenLayer
DELAY=1000

# Sync interval in ms for the aggregator to run internal tasks
SYNC_INTERVAL=7200000

# Metrics port, default 6060
METRICS_PORT=6060

# Metrics export url
METRICS_EXPORT_URL=

# Persistent Storage Path
DATA_DIR_PATH=/home/sky/trigg3rX/triggerx-othentic/data/p2p

othentic-cli node aggregator \
    --json-rpc \
    --internal-tasks \
    --metrics \
    --delay $DELAY \
    --sync-interval $SYNC_INTERVAL \
    --keystore .keystore/aggregator.json \
    --json-rpc.custom-message-enabled