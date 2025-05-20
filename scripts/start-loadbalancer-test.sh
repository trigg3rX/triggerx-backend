#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Docker compose file
COMPOSE_FILE="docker-compose.loadbalancer.yaml"

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
    docker compose -f $COMPOSE_FILE up -d redis
    sleep 5
  fi

  # Check if Redis is accessible
  log_info "Testing Redis connection..."
  if ! docker exec -i triggerx-redis redis-cli ping | grep -q "PONG"; then
    log_error "Cannot connect to Redis. Please check Redis container."
  fi
  
  log_success "Redis is running and accessible"
}

# Start load balancer
start_loadbalancer() {
  log_info "Starting load balancer..."
  docker compose -f $COMPOSE_FILE up -d loadbalancer
  
  # Wait for load balancer to be ready
  log_info "Waiting for load balancer to be ready..."
  while ! curl -s http://localhost:8080/health > /dev/null; do
    echo -n "."
    sleep 1
  done
  echo ""
  log_success "Load balancer is ready"
}

# Start manager containers
start_managers() {
  log_info "Starting manager containers..."
  
  # Start manager1
  log_info "Starting manager1..."
  docker compose -f $COMPOSE_FILE up -d manager1
  
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
  docker compose -f $COMPOSE_FILE up -d manager2
  
  # Wait for manager2 to be ready
  log_info "Waiting for manager2 to be ready..."
  while ! curl -s http://localhost:6062/health > /dev/null; do
    echo -n "."
    sleep 1
  done
  echo ""
  log_success "manager2 is ready"
}

# Register managers with load balancer
register_managers() {
  log_info "Registering managers with load balancer..."
  
  # Register manager1
  curl -X POST http://localhost:8080/register \
    -H "Content-Type: application/json" \
    -d '{
      "id": "manager1",
      "address": "manager1:6061",
      "max_tasks": 100
    }'
  
  # Register manager2
  curl -X POST http://localhost:8080/register \
    -H "Content-Type: application/json" \
    -d '{
      "id": "manager2",
      "address": "manager2:6062",
      "max_tasks": 100
    }'
  
  log_success "Managers registered with load balancer"
}

# Test load balancer
test_loadbalancer() {
  log_info "Testing load balancer..."
  
  # Send a test job request
  curl -X POST http://localhost:8080/job/create \
    -H "Content-Type: application/json" \
    -d '{
      "job_id": 1,
      "task_definition_id": 1,
      "time_interval": 60,
      "time_frame": "1h"
    }'
  
  log_success "Test job request sent"
}

# Main function
main() {
  # Check Redis
  check_redis
  
  # Start load balancer
  start_loadbalancer
  
  # Start managers
  start_managers
  
  # Register managers
  register_managers
  
  # Test load balancer
  test_loadbalancer
  
  log_success "Setup complete! You can now test the load balancer."
  log_info "Load Balancer: http://localhost:8080"
  log_info "Manager 1: http://localhost:6061"
  log_info "Manager 2: http://localhost:6062"
  
  # Keep script running until interrupted
  while true; do
    sleep 10
    
    # Check if all services are still running
    if ! curl -s http://localhost:8080/health > /dev/null || \
       ! curl -s http://localhost:6061/health > /dev/null || \
       ! curl -s http://localhost:6062/health > /dev/null; then
      log_error "One or more services have stopped. Please check the logs."
    fi
  done
}

# Trap Ctrl+C to stop services gracefully
trap 'docker compose -f $COMPOSE_FILE down' INT

# Run main function
main 