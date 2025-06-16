#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 -n <service> -v <version> -p <port>"
    echo "Example: $0 -n health -v 0.0.1 -p 8080"
    exit 1
}

# Parse command-line arguments
while getopts ":n:v:p:" opt; do
    case ${opt} in
        n )
            SERVICE=$OPTARG
            ;;
        v )
            VERSION=$OPTARG
            ;;
        p )
            PORT=$OPTARG
            ;;
        \? )
            echo "Invalid option: $OPTARG" 1>&2
            usage
            ;;
        : )
            echo "Invalid option: $OPTARG requires an argument" 1>&2
            usage
            ;;
    esac
done

# Check if name is provided
if [ -z "$SERVICE" ]; then
    echo "Error: Service is required" 1>&2
    usage
fi

# Check if version is provided
if [ -z "$VERSION" ]; then
    echo "Error: Version is required (e.g., 0.0.1)" 1>&2
    usage
fi

# Check if port is provided
if [ -z "$PORT" ]; then
    echo "Error: Port is required" 1>&2
    usage
fi

# Validate the service from the list of allowed services
if [[ ! "$SERVICE" =~ ^(dbserver|registrar|health|redis|schedulers/time|schedulers/event|schedulers/condition)$ ]]; then
    echo "Error: Invalid service. Allowed services are: dbserver, registrar, health, redis, schedulers/time, schedulers/event, schedulers/condition" 1>&2
    exit 1
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

# Check if container already exists and remove it
if docker ps -a | grep -q "triggerx-${SERVICE}"; then
    echo "Stopping and removing existing container triggerx-${SERVICE}..."
    docker stop triggerx-${SERVICE} 2>/dev/null
    docker rm triggerx-${SERVICE} 2>/dev/null
fi

# Pull the image
# echo "Pulling trigg3rx/triggerx-${SERVICE}:${VERSION}..."
# docker pull trigg3rx/triggerx-${SERVICE}:${VERSION}

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Warning: .env file not found. Container will run without environment variables."
    ENV_FILE=""
else
    ENV_FILE="--env-file .env"
fi

# Run the container
echo "Starting container triggerx-${SERVICE}..."
docker run -d \
    --name triggerx-${SERVICE} \
    ${ENV_FILE} \
    -p ${PORT}:${PORT} \
    --restart unless-stopped \
    trigg3rx/triggerx-${SERVICE}:${VERSION}

# Check if container started successfully
if [ $? -eq 0 ]; then
    echo "Container started successfully!"
    echo "Container ID: $(docker ps -q -f name=triggerx-${SERVICE})"
    echo "Port: ${PORT}"
    echo "To view logs: docker logs triggerx-${SERVICE}"
    echo "To stop: docker stop triggerx-${SERVICE}"
else
    echo "Error: Failed to start container"
    exit 1
fi
