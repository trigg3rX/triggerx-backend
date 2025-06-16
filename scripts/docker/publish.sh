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
if [[ ! "$SERVICE" =~ ^(keeper|dbserver|registrar|health|redis|schedulers/time|schedulers/event|schedulers/condition)$ ]]; then
    echo "Error: Invalid service. Allowed services are: keeper, dbserver, registrar, health, redis, schedulers/time, schedulers/event, schedulers/condition" 1>&2
    exit 1
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

# Login to Docker Hub
docker login

# Push both version and latest tags
docker push trigg3rx/triggerx-${SERVICE}:${VERSION}
docker push trigg3rx/triggerx-${SERVICE}:latest

echo "Successfully tagged and pushed: triggerx-${SERVICE}:${VERSION} and latest tag"
