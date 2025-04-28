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

# Check if Redis is running
check_redis() {
  log_info "Checking if Redis is running..."
  if ! docker ps | grep -q "triggerx-redis"; then
    log_info "Redis container not found, starting it..."
    docker-compose up -d redis
    sleep 5
  fi

  # Check if Redis is accessible
  log_info "Testing Redis connection..."
  if ! docker exec -i triggerx-redis redis-cli ping | grep -q "PONG"; then
    log_error "Cannot connect to Redis. Please check Redis container."
  fi
  
  log_success "Redis is running and accessible"
}

# Helper function to run manager with the right environment
run_manager() {
  local manager_id=$1
  local manager_port=$((8080 + $manager_id))
  local other_port=$((8080 + (3 - $manager_id)))  # If MANAGER_ID is 1, OTHER_PORT is 8082, and vice versa
  
  log_info "Starting Manager $manager_id on port $manager_port..."
  
  # Set environment variables
  export MANAGER_RPC_PORT=$manager_port
  export MANAGER_HA_ENABLED=true
  export REDIS_ADDRESS=localhost:6379
  export OTHER_MANAGER_ADDRESSES=localhost:$other_port
  
  # Run the manager in the background
  go run ./cmd/manager/main.go &
  
  # Save the process ID for later termination
  echo $! > /tmp/manager-$manager_id.pid
  
  log_success "Manager $manager_id started with PID $(cat /tmp/manager-$manager_id.pid)"
}

# Stop running managers
stop_managers() {
  log_info "Stopping any running managers..."
  
  # Stop Manager 1 if running
  if [ -f /tmp/manager-1.pid ]; then
    pid=$(cat /tmp/manager-1.pid)
    if ps -p $pid > /dev/null; then
      kill $pid
      log_info "Stopped manager 1 (PID $pid)"
    fi
    rm /tmp/manager-1.pid
  fi
  
  # Stop Manager 2 if running
  if [ -f /tmp/manager-2.pid ]; then
    pid=$(cat /tmp/manager-2.pid)
    if ps -p $pid > /dev/null; then
      kill $pid
      log_info "Stopped manager 2 (PID $pid)"
    fi
    rm /tmp/manager-2.pid
  fi
  
  # Kill any remaining processes
  pkill -f "go run ./cmd/manager/main.go" || true
  
  log_success "All managers stopped"
}

# Helper function to wait for manager to be ready
wait_for_manager() {
  local manager_id=$1
  local port=$((8080 + $manager_id))
  local max_retries=30
  local retries=0
  
  log_info "Waiting for Manager $manager_id to be ready..."
  
  while ! curl -s http://localhost:$port/health > /dev/null && [ $retries -lt $max_retries ]; do
    retries=$((retries + 1))
    echo -n "."
    sleep 1
  done
  
  echo ""
  
  if [ $retries -ge $max_retries ]; then
    log_error "Manager $manager_id did not become ready in time"
  fi
  
  log_success "Manager $manager_id is ready"
}

# Trap Ctrl+C to stop managers gracefully
trap stop_managers INT

# Main function
main() {
  # Stop any existing managers
  stop_managers
  
  # Start Redis if needed
  check_redis
  
  # Start manager 1
  run_manager 1
  
  # Wait for manager 1 to be ready
  wait_for_manager 1
  
  # Start manager 2
  run_manager 2
  
  # Wait for manager 2 to be ready
  wait_for_manager 2
  
  log_success "Both managers are running successfully!"
  log_info "Manager 1: http://localhost:8081"
  log_info "Manager 2: http://localhost:8082"
  log_info "Press Ctrl+C to stop all managers"
  
  # Keep script running until interrupted
  while true; do
    sleep 10
    
    # Check if managers are still running
    pid1=$(cat /tmp/manager-1.pid 2>/dev/null)
    pid2=$(cat /tmp/manager-2.pid 2>/dev/null)
    
    if ! ps -p $pid1 > /dev/null || ! ps -p $pid2 > /dev/null; then
      log_error "One or both managers have stopped. Please check the logs."
    fi
  done
}

# Run main function
main 