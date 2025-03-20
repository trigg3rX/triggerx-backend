#! /bin/bash

docker build -t keeper-execution .

sleep 1

docker tag keeper-execution trigg3rx/keeper-execution:latest

sleep 1

docker login

sleep 2

docker push trigg3rx/keeper-execution:latest