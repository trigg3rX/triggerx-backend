#! /bin/bash

source .env

docker pull golang

go run cmd/keeper/main.go
