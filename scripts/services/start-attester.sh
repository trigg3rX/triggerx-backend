#!/bin/bash

source .env

# othentic-cli node attester \
#     /ip4/127.0.0.1/tcp/9876/p2p/12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB \
#     --avs-webapi http://127.0.0.1 \
#     --avs-webapi-port 9005 \
#     --json-rpc \
#     --json-rpc.port 9006 \
#     --json-rpc.custom-message-enabled \
#     --p2p.port 33333 \
#     --p2p.datadir data/peerstore/luffy \
#     --p2p.discovery-interval 10000 \
#     --metrics \
#     --metrics.port 9009

othentic-cli node attester \
    /ip4/127.0.0.1/tcp/9876/p2p/12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB \
    --avs-webapi http://127.0.0.1 \
    --avs-webapi-port 9015 \
    --json-rpc \
    --json-rpc.port 9016 \
    --json-rpc.custom-message-enabled \
    --p2p.port 33332 \
    --p2p.datadir data/peerstore/zoro \
    --p2p.discovery-interval 10000 \
    --metrics \
    --metrics.port 9019