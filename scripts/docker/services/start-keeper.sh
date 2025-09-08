#! /bin/bash

# Source environment variables if .env exists
if [ -f .env ]; then
    source .env
fi

# Ensure mounted directories exist and have proper permissions
mkdir -p data/cache
mkdir -p data/logs
mkdir -p data/peerstore

# Set up signal handlers
trap cleanup SIGTERM SIGINT SIGQUIT

# Start the keeper service
./triggerx-keeper
