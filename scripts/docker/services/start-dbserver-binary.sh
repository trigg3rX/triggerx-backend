#!/bin/bash

# Load environment variables
if [ -f .env ]; then
    source .env
fi

# Set default environment variables if not set
export GIN_MODE=${GIN_MODE:-release}
export LOG_LEVEL=${LOG_LEVEL:-info}

# Print startup information
echo "Starting ${SERVICE} service..."
echo "Environment: ${GIN_MODE}"
echo "Log Level: ${LOG_LEVEL}"

# Run the service
exec /app/dbserver