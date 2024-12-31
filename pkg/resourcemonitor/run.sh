#!/bin/bash

echo "Building and running resource monitor..."

# Build and run the container with log volume
docker build -t resource-monitor .
docker run --rm \
    --memory="512m" \
    --memory-swap="512m" \
    -v "$(pwd)/logs:/app/logs" \
    resource-monitor

echo "Resource monitoring complete. Logs are available in the logs directory."