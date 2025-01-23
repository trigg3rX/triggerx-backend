#!/bin/bash

docker stop triggerx-keeper
docker rm triggerx-keeper

# Build the image
docker build -t triggerx-keeper -f keeper/Dockerfile .

# Run the container
docker run -d \
    --name triggerx-keeper \
    triggerx-keeper