#!/bin/bash

# Check if version argument is provided
# if [ -z "$1" ]; then
#     echo "Error: Version argument is required"
#     echo "Usage: $0 <version>"
#     exit 1
# fi

# VERSION=$1

# # Build the binary
# ./scripts/binary/build.sh

# # Build the Zip file
# tar -czf ./release/triggerx-backend_${VERSION}_linux_amd64.tar.gz -C ~/bin triggerx

# # Create the checksum file
# sha256sum ./release/triggerx-backend_${VERSION}_linux_amd64.tar.gz > ./release/triggerx-backend_${VERSION}_checksums.txt
