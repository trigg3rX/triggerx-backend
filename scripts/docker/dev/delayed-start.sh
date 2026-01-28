#!/bin/sh
# Wrapper script to add a delay before starting a service
# Usage: delayed-start.sh <delay_seconds> <command> [args...]

DELAY=$1
shift

if [ -n "$DELAY" ] && [ "$DELAY" -gt 0 ]; then
    sleep "$DELAY"
fi

exec "$@"
