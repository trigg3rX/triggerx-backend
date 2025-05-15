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
    docker compose up -d redis
    sleep 5
  fi

  # Check if Redis is accessible
  log_info "Testing Redis connection..."
  if ! docker exec -i triggerx-redis redis-cli ping | grep -q "PONG"; then
    log_error "Cannot connect to Redis. Please check Redis container."
  fi
  
  log_success "Redis is running and accessible"
}

# Start manager containers
start_managers() {
  log_info "Starting manager containers..."
  
  # Start manager1
  log_info "Starting manager1..."
  docker compose up -d manager1
  
  # Wait for manager1 to be ready
  log_info "Waiting for manager1 to be ready..."
  while ! curl -s http://localhost:6061/health > /dev/null; do
    echo -n "."
    sleep 1
  done
  echo ""
  log_success "manager1 is ready"
  
  # Start manager2
  log_info "Starting manager2..."
  docker compose up -d manager2
  
  # Wait for manager2 to be ready
  log_info "Waiting for manager2 to be ready..."
  while ! curl -s http://localhost:6062/health > /dev/null; do
    echo -n "."
    sleep 1
  done
  echo ""
  log_success "manager2 is ready"
}

# Stop manager containers
stop_managers() {
  log_info "Stopping manager containers..."
  docker compose stop manager1 manager2
  log_success "Manager containers stopped"
}

# Check manager status
check_managers() {
  log_info "Checking manager status..."
  
  # Check manager1
  if curl -s http://localhost:6061/status | grep -q "ok"; then
    log_success "manager1 is running"
  else
    log_error "manager1 is not running properly"
  fi
  
  # Check manager2
  if curl -s http://localhost:6062/status | grep -q "ok"; then
    log_success "manager2 is running"
  else
    log_error "manager2 is not running properly"
  fi
}

# Main function
main() {
  # Stop any existing managers
  stop_managers
  
  # Start Redis if needed
  check_redis
  
  # Start managers
  start_managers
  
  # Check manager status
  check_managers
  
  log_success "Both managers are running successfully!"
  log_info "Manager 1: http://localhost:6061"
  log_info "Manager 2: http://localhost:6062"
  log_info "Press Ctrl+C to stop all managers"
  
  # Keep script running until interrupted
  while true; do
    sleep 10
    
    # Check if managers are still running
    if ! curl -s http://localhost:6061/health > /dev/null || ! curl -s http://localhost:6062/health > /dev/null; then
      log_error "One or both managers have stopped. Please check the logs."
    fi
  done
}

# Trap Ctrl+C to stop managers gracefully
trap stop_managers INT

# Run main function
main 