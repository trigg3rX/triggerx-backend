#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Add configuration for remote servers
MANAGER1_HOST="SERVER1_PUBLIC_IP"
MANAGER2_HOST="SERVER2_PUBLIC_IP"
MANAGER1_PORT="6061"
MANAGER2_PORT="6062"

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

# Check Redis connection
check_redis() {
  log_info "Checking Redis connection..."
  if ! redis-cli -h $REDIS_HOST ping | grep -q "PONG"; then
    log_error "Cannot connect to Redis at $REDIS_HOST"
  fi
  log_success "Redis is accessible"
}

# Check both managers' health
check_managers() {
  log_info "Checking health of both managers..."
  
  # Check Manager 1
  if curl -s http://${MANAGER1_HOST}:${MANAGER1_PORT}/health > /dev/null; then
    log_success "Manager 1 is healthy"
  else
    log_error "Manager 1 is not responding"
  fi
  
  # Check Manager 2
  if curl -s http://${MANAGER2_HOST}:${MANAGER2_PORT}/health > /dev/null; then
    log_success "Manager 2 is healthy"
  else
    log_error "Manager 2 is not responding"
  fi
  
  # Check cross-communication
  log_info "Checking manager cross-communication..."
  
  # Check if Manager 1 can see Manager 2
  if curl -s http://${MANAGER1_HOST}:${MANAGER1_PORT}/peer-status | grep -q "${MANAGER2_HOST}"; then
    log_success "Manager 1 can see Manager 2"
  else
    log_error "Manager 1 cannot communicate with Manager 2"
  fi
  
  # Check if Manager 2 can see Manager 1
  if curl -s http://${MANAGER2_HOST}:${MANAGER2_PORT}/peer-status | grep -q "${MANAGER1_HOST}"; then
    log_success "Manager 2 can see Manager 1"
  else
    log_error "Manager 2 cannot communicate with Manager 1"
  fi
}

# Main monitoring loop
monitor_cluster() {
  while true; do
    check_redis
    check_managers
    
    log_info "Cluster health check completed. Waiting 30 seconds..."
    sleep 30
  done
}

# Run the monitoring
monitor_cluster
