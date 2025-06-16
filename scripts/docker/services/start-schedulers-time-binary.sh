#!/bin/bash

# Load environment variables
if [ -f .env ]; then
    source .env
fi

# Run the service
exec /app/schedulers/time