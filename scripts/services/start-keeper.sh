#! /bin/bash

source .env

echo "Starting Keeper node ..."

./triggerx-keeper &
KEEPER_PID=$!

# Handle shutdown signals to properly terminate both processes
trap "kill $KEEPER_PID; exit" SIGTERM SIGINT

# Wait for either process to exit
wait -n