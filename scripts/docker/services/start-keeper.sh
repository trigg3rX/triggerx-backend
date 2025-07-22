#! /bin/bash

# Source environment variables if .env exists
if [ -f .env ]; then
    source .env
fi

# Ensure mounted directories exist and have proper permissions
mkdir -p data/cache
mkdir -p data/logs
mkdir -p data/peerstore/attester

# Set permissions for mounted directories
chmod -R 755 data

# Pull the latest golang image for container execution
docker pull golang:1.21-alpine

# Start the keeper service
./triggerx-keeper
