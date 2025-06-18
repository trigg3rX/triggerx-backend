#! /bin/bash

source .env

docker pull golang

./triggerx-keeper
