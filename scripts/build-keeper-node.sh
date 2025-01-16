#!/bin/bash

# Set variables
DOCKER_USERNAME="praptishah"
IMAGE_NAME="triggerx-keeper"
FULL_IMAGE_NAME="$DOCKER_USERNAME/$IMAGE_NAME:latest"

# Generate cosign key pair if not exists
if [ ! -f cosign.key ]; then
    cosign generate-key-pair
fi

# Build the image
docker build -t $FULL_IMAGE_NAME -f keeper/Dockerfile .

# Push to Docker Hub
docker push $FULL_IMAGE_NAME

# Sign the image
cosign sign --key cosign.key $FULL_IMAGE_NAME

# Verify (optional)
cosign verify --key cosign.pub $FULL_IMAGE_NAME

# Run the container
docker run -d \
    --name triggerx-keeper \
    $FULL_IMAGE_NAME