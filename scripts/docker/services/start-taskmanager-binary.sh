#!/bin/bash

# Set umask to ensure log files are group-readable (664 instead of 600)
umask 002

# Load environment variables from mounted .env file
if [ -f /home/appuser/.env ]; then
    echo "Loading environment variables from .env file..."
    source /home/appuser/.env
else
    echo "No .env file found, proceeding with default environment..."
fi

# Run the service
exec /home/appuser/taskmanager