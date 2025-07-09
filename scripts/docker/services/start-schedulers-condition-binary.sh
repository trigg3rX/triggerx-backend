#!/bin/bash

# Load environment variables from mounted .env file
if [ -f /home/appuser/.env ]; then
    echo "Loading environment variables from .env file..."
    source /home/appuser/.env
else
    echo "No .env file found, proceeding with default environment..."
fi

# Run the service
exec /home/appuser/schedulers-condition