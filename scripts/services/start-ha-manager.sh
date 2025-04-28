#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_info() {
  echo -e "${YELLOW}[INFO]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
  exit 1
}

# Debug function
debug_info() {
  echo "=== Debug Information ==="
  echo "Current directory: $(pwd)"
  echo "Environment variables:"
  env | grep -E 'MANAGER|REDIS|NODE'
  echo "Files in current directory:"
  ls -la
  echo "========================="
}

# Check Redis connectivity
check_redis() {
  log_info "Checking Redis connectivity..."
  
  max_retries=30
  count=0
  
  while ! redis-cli -h "${REDIS_HOST:-redis}" -p "${REDIS_PORT:-6379}" ping; do
    count=$((count + 1))
    if [ $count -ge $max_retries ]; then
      log_error "Failed to connect to Redis after $max_retries attempts"
    fi
    log_info "Waiting for Redis... (attempt $count/$max_retries)"
    sleep 1
  done
  
  log_success "Successfully connected to Redis"
}

# Check if other manager nodes are accessible
check_other_nodes() {
  if [ -z "${OTHER_MANAGER_ADDRESSES}" ]; then
    log_info "No other manager nodes specified, running in standalone mode"
    return
  fi
  
  log_info "Checking connectivity to other manager nodes..."
  
  IFS=',' read -ra NODES <<< "${OTHER_MANAGER_ADDRESSES}"
  for node in "${NODES[@]}"; do
    host=$(echo "$node" | cut -d: -f1)
    port=$(echo "$node" | cut -d: -f2)
    
    if ! nc -z "$host" "$port"; then
      log_info "Manager node $node is not yet available"
    else
      log_info "Manager node $node is accessible"
    fi
  done
}

# Start the manager with proper environment
start_manager() {
  # Ensure required environment variables
  : "${MANAGER_RPC_PORT:?Need to set MANAGER_RPC_PORT}"
  : "${NODE_ID:?Need to set NODE_ID}"
  
  export MANAGER_HA_ENABLED=true
  export CONFIG_FILE="/app/config/ha_config.yaml"
  
  log_info "Starting manager node $NODE_ID on port $MANAGER_RPC_PORT"
  
  # Print debug information
  debug_info
  
  # Check if the manager binary exists
  if [ ! -f "./triggerx-manager" ]; then
    log_error "Manager binary not found at ./triggerx-manager"
  fi
  
  # Run the manager
  log_info "Running manager with command: ./triggerx-manager"
  exec ./triggerx-manager
}

# Main execution
main() {
  # Check dependencies
  check_redis
  check_other_nodes
  
  # Start manager
  start_manager
}

main "$@"