#!/bin/bash

# Build the Docker image
docker build -t triggerx-keeper -f keeper/Dockerfile .

# Run the container
# -d: run in detached mode (background)
# --name: give the container a name for easy reference
# --restart: automatically restart the container if it crashes
docker run -d \
    --name triggerx-keeper \
    triggerx-keeper