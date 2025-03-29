#!/bin/sh
echo "Starting keeper node..."

# Start keeper-execution in the background
./keeper-execution &
KEEPER_PID=$!

# Start othentic-cli in the background
echo "Starting othentic-cli..."
othentic-cli node attester \
    /ip4/157.173.218.229/tcp/9876/p2p/12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB \
    --metrics \
    --avs-webapi http://127.0.0.1 \
    --avs-webapi-port 9005 \
    --json-rpc \
    --json-rpc.port 9006 \
    --json-rpc.custom-message-enabled &
OTHENTIC_PID=$!

# Handle shutdown signals to properly terminate both processes
trap "kill $KEEPER_PID $OTHENTIC_PID; exit" SIGTERM SIGINT

# Wait for either process to exit
wait -n

# If one process exits, kill the other and exit with the same code
EXIT_CODE=$?
kill $KEEPER_PID $OTHENTIC_PID 2>/dev/null
exit $EXIT_CODE