#!/bin/sh

source .env

echo "Starting keeper node..."

# Start keeper-execution in the background
./keeper-execution &
KEEPER_PID=$!

# Start othentic-cli in the background
# echo "Starting othentic-cli..."
# othentic-cli node attester \
#     /ip4/157.173.218.229/tcp/9876/p2p/$OTHENTIC_BOOTSTRAP_ID \
#     --avs-webapi http://127.0.0.1 \
#     --avs-webapi-port $OPERATOR_RPC_PORT \
#     --json-rpc.custom-message-enabled \
#     --p2p.port $OPERATOR_P2P_PORT \
#     --p2p.datadir data/peerstore/attester \
#     --p2p.discovery-interval 60000 \
#     --metrics \
#     --metrics.port $OPERATOR_METRICS_PORT &
# OTHENTIC_PID=$!

# Start othentic-cli in the background
echo "Starting othentic-cli..."
othentic-cli node attester \
    /ip4/127.0.0.1/tcp/9876/p2p/$OTHENTIC_BOOTSTRAP_ID \
    --avs-webapi http://127.0.0.1 \
    --avs-webapi-port $OPERATOR_RPC_PORT \
    --json-rpc.custom-message-enabled \
    --p2p.port $OPERATOR_P2P_PORT \
    --p2p.datadir data/peerstore/attester \
    --p2p.discovery-interval 60000 \
    --metrics \
    --metrics.port $OPERATOR_METRICS_PORT &
OTHENTIC_PID=$!

# Handle shutdown signals to properly terminate both processes
trap "kill $KEEPER_PID $OTHENTIC_PID; exit" SIGTERM SIGINT

# Wait for either process to exit
wait -n

# If one process exits, kill the other and exit with the same code
EXIT_CODE=$?
kill $KEEPER_PID $OTHENTIC_PID 2>/dev/null
exit $EXIT_CODE