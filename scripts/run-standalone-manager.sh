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

# Helper function to run manager with the right environment
run_standalone_manager() {
  local port=${1:-6061}
  
  log_info "Starting Standalone Manager on port $port..."
  
  # Set environment variables
  export MANAGER_RPC_PORT=$port
  export MANAGER_HA_ENABLED=false  # Disable HA mode
  
  # Unset any HA-related variables to avoid confusion
  unset REDIS_ADDRESS
  unset OTHER_MANAGER_ADDRESSES
  
  # Run the manager 
  log_info "Running manager in standalone mode - you will see logs directly in this console"
  log_info "Press Ctrl+C to stop"
  
  # Run the manager in the foreground
  go run ./cmd/manager/main.go
}

# Main function
main() {
  local port=${1:-6061}
  
  # Check if port is a valid number
  if ! [[ "$port" =~ ^[0-9]+$ ]]; then
    log_error "Invalid port number: $port"
  fi
  
  # Kill any existing manager processes
  log_info "Stopping any existing manager processes..."
  pkill -f "go run ./cmd/manager/main.go" || true
  
  # Run the standalone manager
  run_standalone_manager $port
}

# Run main function with port argument if provided
main $1