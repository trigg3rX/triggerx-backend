#! /bin/bash

# Default to foreground (normal mode)
BACKGROUND=false

# Parse arguments
while getopts ":p" opt; do
  case ${opt} in
    p)
      BACKGROUND=true
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

if [ "$BACKGROUND" = true ]; then
    echo "Starting Health Service in background (logs hidden)..."
    go run ./cmd/health/main.go > /dev/null 2>&1 &
else
    # Run normally (foreground, logs visible)
    go run ./cmd/health/main.go
fi
