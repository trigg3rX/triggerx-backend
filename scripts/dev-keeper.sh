#!/bin/bash

# Function to display usage
usage() {
    echo "Usage: $0 -v <version>"
    echo "Example: $0 -v 0.0.1"
    exit 1
}

# Parse command-line arguments
while getopts ":v:" opt; do
    case ${opt} in
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

# Check if version is provided
if [ -z "$VERSION" ]; then
    echo "Error: Version is required" 1>&2
    usage
fi

# Validate version format (basic regex for semantic versioning)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 0.0.1)" 1>&2
    exit 1
fi

# Build Docker image with specified version
docker build -t dev-keeper:${VERSION} .

# Tag images
docker tag dev-keeper:${VERSION} trigg3rx/dev-keeper:${VERSION}
docker tag dev-keeper:${VERSION} trigg3rx/dev-keeper:latest

# Login to Docker Hub
docker login

# Push version tags
docker push trigg3rx/dev-keeper:${VERSION}
docker push trigg3rx/dev-keeper:latest

echo "Successfully built and pushed version: $VERSION and latest tag"