#!/bin/bash

docker rm -f triggerx-keeper || true

docker build -t triggerx-keeper -f keeper/Dockerfile .

docker run -d --name triggerx-keeper triggerx-keeper