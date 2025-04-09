#!/bin/bash

source .env

othentic-cli node attester \
    /ip4/127.0.0.1/tcp/9876/p2p/12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB \
    --avs-webapi http://127.0.0.1 \
    --avs-webapi-port 9003 \
    --json-rpc.custom-message-enabled \
    --p2p.port 33339 \
    --p2p.datadir data/peerstore/nexus \
    --p2p.discovery-interval 10000