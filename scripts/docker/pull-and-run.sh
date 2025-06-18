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

# Convert service name to Docker-compatible name
DOCKER_NAME=$(echo $SERVICE | sed 's/\//-/g')

# Check if container already exists and remove it
if docker ps -a | grep -q "triggerx-${DOCKER_NAME}"; then
    echo "Stopping and removing existing container triggerx-${DOCKER_NAME}..."
    docker stop triggerx-${DOCKER_NAME} 2>/dev/null
    docker rm triggerx-${DOCKER_NAME} 2>/dev/null
fi

# Check if local image exists
if ! docker images --filter reference="triggerx-${DOCKER_NAME}:${VERSION}" | grep "triggerx-${DOCKER_NAME}"; then
    echo "Local image triggerx-${DOCKER_NAME}:${VERSION} does not exist" 1>&2
    IMAGE_NAME="trigg3rx/triggerx-${DOCKER_NAME}:${VERSION}"
    echo "Pulling ${IMAGE_NAME}..."
    docker pull ${IMAGE_NAME}
else
    echo "Using local image triggerx-${DOCKER_NAME}:${VERSION}"
    IMAGE_NAME="triggerx-${DOCKER_NAME}:${VERSION}"
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Warning: .env file not found. Container will run without environment variables."
    ENV_FILE=""
else
    ENV_FILE="-v ./.env:/root/.env"
fi

if [[ "$SERVICE" == "registrar" ]]; then
    # Run the container
    echo "Starting container triggerx-${DOCKER_NAME}..."
    docker run -d \
        --name triggerx-${DOCKER_NAME} \
        ${ENV_FILE} \
        -v ./pkg/bindings/abi:/root/pkg/bindings/abi \
        -p ${PORT}:${PORT} \
        --restart unless-stopped \
        ${IMAGE_NAME}
else
    # Run the container
    echo "Starting container triggerx-${DOCKER_NAME}..."
    docker run -d \
        --name triggerx-${DOCKER_NAME} \
        ${ENV_FILE} \
        --network host \
        --restart unless-stopped \
        ${IMAGE_NAME}
fi

# Check if container started successfully
if [ $? -eq 0 ]; then
    echo "Container started successfully!"
    echo "Container ID: $(docker ps -q -f name=triggerx-${DOCKER_NAME})"
    echo "Port: ${PORT}"
    echo "To view logs: docker logs triggerx-${DOCKER_NAME}"
    echo "To stop: docker stop triggerx-${DOCKER_NAME}"
else
    echo "Error: Failed to start container"
    exit 1
fi
