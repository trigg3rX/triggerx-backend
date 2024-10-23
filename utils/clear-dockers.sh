#!/bin/bash

docker stop $(docker ps -q) || true
docker rm $(docker ps -aq) || true
