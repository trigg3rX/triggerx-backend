#!/bin/bash

# Build the Docker image
docker build -t triggerx-keeper -f keeper/Dockerfile .

# Create a Docker network if it doesn't exist
docker network create triggerx-network || true

# Get the host's IP address
HOST_IP=$(ip -4 addr show docker0 | grep -Po 'inet \K[\d.]+')

# Start 4 keeper containers
for i in {1..4}
do
    docker stop triggerx-keeper-$i || true
    docker rm -f triggerx-keeper-$i || true
    docker run -d --name triggerx-keeper-$i \
                --network triggerx-network \
                -e KEEPER_ID=$i \
                -e HOST_IP=$HOST_IP \
                -p 300$(($i)):3000 \
                triggerx-keeper
done