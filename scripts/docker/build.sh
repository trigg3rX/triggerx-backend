#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 -n <service> -v <version>"
    echo "Example: $0 -n keeper -v 0.0.1"
    exit 1
}

# Parse command-line arguments
while getopts ":n:v:" opt; do
    case ${opt} in
        n )
            SERVICE=$OPTARG
            ;;
        v )
            VERSION=$OPTARG
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

# Validate the service from the list of allowed services
if [[ ! "$SERVICE" =~ ^(keeper|dbserver|registrar|health|redis|schedulers/time|schedulers/condition|all)$ ]]; then
    echo "Error: Invalid service. Allowed services are: keeper, dbserver, registrar, health, redis, schedulers/time, schedulers/condition" 1>&2
    exit 1
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

if [[ "$SERVICE" == "all" ]]; then
    # Build all services
    # for service in dbserver registrar health redis schedulers/time schedulers/condition; do
    for service in schedulers/time schedulers/condition; do
        # Convert service name to Docker-compatible name
        DOCKER_NAME=$(echo $service | sed 's/\//-/g')

        echo "Building $service..."
        docker build --no-cache \
            -f Dockerfile.backend \
            --build-arg SERVICE=${service} \
            --build-arg DOCKER_NAME=${DOCKER_NAME} \
            -t triggerx-${DOCKER_NAME}:${VERSION} .
    done
elif [[ "$SERVICE" == "keeper" ]]; then
    echo "Building $SERVICE..."
    docker build --no-cache \
        -f Dockerfile.keeper \
        -t triggerx-keeper:${VERSION} .
else
    echo "Building $SERVICE..."
    # Convert service name to Docker-compatible name
    DOCKER_NAME=$(echo $SERVICE | sed 's/\//-/g')

    echo "DOCKER_NAME: $DOCKER_NAME"
    # Build a single service
    echo "Building $SERVICE..."
    docker build --no-cache \
        -f Dockerfile.backend \
        --build-arg SERVICE=${SERVICE} \
        --build-arg DOCKER_NAME=${DOCKER_NAME} \
        -t triggerx-${DOCKER_NAME}:${VERSION} .
fi

echo "Successfully built: triggerx-${SERVICE}:${VERSION} and latest"
