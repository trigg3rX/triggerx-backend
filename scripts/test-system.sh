#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
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

log_warning() {
  echo -e "${BLUE}[WARNING]${NC} $1"
}

# Check if docker is running
check_docker() {
  log_info "Checking if Docker is running..."
  if ! docker info &>/dev/null; then
    log_error "Docker is not running. Please start Docker and try again."
  fi
  log_success "Docker is running"
}

# Start containers if needed
start_containers() {
  log_info "Checking if docker-compose services need to be started..."
  
  # Check if any container is running
  if ! docker ps | grep -q "triggerx"; then
    log_warning "No triggerx containers are running. Starting with docker-compose..."
    docker-compose up -d
    
    # Wait for containers to be ready
    log_info "Waiting for containers to be ready..."
    sleep 20
  else
    # Check individual containers and start missing ones
    local missing_containers=false
    
    if ! docker ps | grep -q "triggerx-scylla-1"; then
      log_warning "ScyllaDB node 1 is not running. Starting it..."
      docker start triggerx-scylla-1
      missing_containers=true
    fi
    
    if ! docker ps | grep -q "triggerx-scylla-2"; then
      log_warning "ScyllaDB node 2 is not running. Starting it..."
      docker start triggerx-scylla-2
      missing_containers=true
    fi
    
    if ! docker ps | grep -q "triggerx-redis"; then
      log_warning "Redis is not running. Starting it..."
      docker start triggerx-redis
      missing_containers=true
    fi
    
    if ! docker ps | grep -q "triggerx-manager-1"; then
      log_warning "Manager 1 is not running. Starting it..."
      docker start triggerx-manager-1
      missing_containers=true
    fi
    
    if ! docker ps | grep -q "triggerx-manager-2"; then
      log_warning "Manager 2 is not running. Starting it..."
      docker start triggerx-manager-2
      missing_containers=true
    fi
    
    # Wait for containers to be ready if any were started
    if $missing_containers; then
      log_info "Waiting for containers to be ready..."
      sleep 15
    fi
  fi
}

# Check if containers are running
check_containers() {
  log_info "Checking if required containers are running..."
  
  local all_running=true
  
  # Check ScyllaDB nodes
  if ! docker ps | grep -q "triggerx-scylla-1"; then
    log_error "ScyllaDB node 1 is not running. Please start docker-compose or run this script with --start flag."
    all_running=false
  fi
  
  if ! docker ps | grep -q "triggerx-scylla-2"; then
    log_error "ScyllaDB node 2 is not running. Please start docker-compose or run this script with --start flag."
    all_running=false
  fi
  
  # Check Redis
  if ! docker ps | grep -q "triggerx-redis"; then
    log_error "Redis is not running. Please start docker-compose or run this script with --start flag."
    all_running=false
  fi
  
  # Check Manager instances
  if ! docker ps | grep -q "triggerx-manager-1"; then
    log_error "Manager 1 is not running. Please start docker-compose or run this script with --start flag."
    all_running=false
  fi
  
  if ! docker ps | grep -q "triggerx-manager-2"; then
    log_error "Manager 2 is not running. Please start docker-compose or run this script with --start flag."
    all_running=false
  fi
  
  if $all_running; then
    log_success "All required containers are running"
  fi
}

# Test ScyllaDB cluster status
test_db_cluster() {
  log_info "Testing ScyllaDB cluster status..."
  
  # Check cluster status
  log_info "Checking cluster status..."
  cluster_status=$(docker exec -i triggerx-scylla-1 nodetool status)
  echo "$cluster_status"
  
  # Verify there are 2 nodes and they're up (UN status)
  node_count=$(echo "$cluster_status" | grep -c "UN")
  if [[ $node_count -ne 2 ]]; then
    log_error "Expected 2 nodes in the cluster, found $node_count"
  fi
  
  log_success "ScyllaDB cluster is healthy with $node_count nodes"
}

# Test ScyllaDB data replication
test_db_replication() {
  log_info "Testing ScyllaDB data replication..."
  
  # Generate unique test ID
  test_id=$(date +%s)
  
  # Insert data through node 1
  log_info "Inserting test data into node 1..."
  docker exec -i triggerx-scylla-1 cqlsh -e "USE triggerx; INSERT INTO user_data (user_id, user_address, created_at) VALUES ($test_id, 'test_replication_$test_id', toTimestamp(now()));"
  
  # Read data from node 2
  log_info "Reading test data from node 2..."
  result=$(docker exec -i triggerx-scylla-2 cqlsh -e "USE triggerx; SELECT * FROM user_data WHERE user_id = $test_id;")
  
  if ! echo "$result" | grep -q "test_replication_$test_id"; then
    log_error "Replication test failed: Data written to node 1 not found in node 2"
  fi
  
  log_success "Data replication between nodes is working"
}

# Test ScyllaDB failover
test_db_failover() {
  log_info "Testing ScyllaDB failover..."
  
  # Generate unique test ID
  test_id=$(date +%s)
  
  # Insert data through node 1
  log_info "Inserting test data into node 1..."
  docker exec -i triggerx-scylla-1 cqlsh -e "USE triggerx; INSERT INTO user_data (user_id, user_address, created_at) VALUES ($test_id, 'test_failover_$test_id', toTimestamp(now()));"
  
  # Stop node 1
  log_info "Stopping node 1 to simulate failure..."
  docker stop triggerx-scylla-1
  
  # Wait for node to stop
  sleep 5
  
  # Try to read data from node 2
  log_info "Reading data from node 2 after node 1 is down..."
  result=$(docker exec -i triggerx-scylla-2 cqlsh -e "USE triggerx; SELECT * FROM user_data WHERE user_id = $test_id;")
  
  if ! echo "$result" | grep -q "test_failover_$test_id"; then
    # Start node 1 before exiting
    docker start triggerx-scylla-1
    sleep 10
    log_error "Failover test failed: Data written to node 1 not accessible from node 2 after node 1 failure"
  fi
  
  # Restart node 1
  log_info "Restarting node 1..."
  docker start triggerx-scylla-1
  sleep 10
  
  log_success "Database failover test passed successfully"
}

# Check manager API
check_manager_api() {
  local port=$1
  log_info "Checking Manager API on port $port..."
  
  # Try different endpoints to see what's working
  echo "Checking root endpoint (/):"
  curl -s -v http://localhost:$port/ 2>&1
  
  echo "Checking /health endpoint:"
  curl -s -v http://localhost:$port/health 2>&1
  
  echo "Checking /status endpoint:"
  curl -s -v http://localhost:$port/status 2>&1
  
  echo "Checking network connectivity to container:"
  docker exec -i triggerx-manager-$((port - 8080)) netstat -tulpn | grep LISTEN
}

# Test Manager health
test_manager_health() {
  log_info "Testing Manager instances health..."
  
  # Debug API access to both managers
  check_manager_api 8081
  check_manager_api 8082
  
  # Check Manager 1
  log_info "Checking Manager 1 health..."
  local manager1_health=$(curl -s -v http://localhost:8081/health 2>&1)
  echo "Manager 1 health response:"
  echo "$manager1_health"
  
  # Try multiple times for Manager 1
  local retry_count=0
  local max_retries=3
  while ! echo "$manager1_health" | grep -q "ok" && [ $retry_count -lt $max_retries ]; do
    log_warning "Manager 1 health check failed, retrying in 10 seconds (attempt $((retry_count+1))/$max_retries)..."
    sleep 10
    manager1_health=$(curl -s -v http://localhost:8081/health 2>&1)
    echo "Manager 1 health retry response:"
    echo "$manager1_health"
    ((retry_count++))
  done
  
  if ! echo "$manager1_health" | grep -q "ok"; then
    log_warning "Checking Manager 1 container logs for errors..."
    docker logs triggerx-manager-1 --tail 20
    log_error "Manager 1 health check failed after $max_retries retries"
  fi
  
  # Check Manager 2
  log_info "Checking Manager 2 health..."
  local manager2_health=$(curl -s -v http://localhost:8082/health 2>&1)
  echo "Manager 2 health response:"
  echo "$manager2_health"
  
  # Try multiple times for Manager 2
  retry_count=0
  while ! echo "$manager2_health" | grep -q "ok" && [ $retry_count -lt $max_retries ]; do
    log_warning "Manager 2 health check failed, retrying in 10 seconds (attempt $((retry_count+1))/$max_retries)..."
    sleep 10
    manager2_health=$(curl -s -v http://localhost:8082/health 2>&1)
    echo "Manager 2 health retry response:"
    echo "$manager2_health"
    ((retry_count++))
  done
  
  if ! echo "$manager2_health" | grep -q "ok"; then
    log_warning "Checking Manager 2 container logs for errors..."
    docker logs triggerx-manager-2 --tail 20
    log_error "Manager 2 health check failed after $max_retries retries"
  fi
  
  log_success "Both manager instances are healthy"
}

# Test Manager leader election
test_manager_leadership() {
  log_info "Testing Manager leader election..."
  
  # Get current leader status
  leader1=$(curl -s http://localhost:8081/status | grep -o '"isLeader":[^,]*' || echo "Failed to get status")
  leader2=$(curl -s http://localhost:8082/status | grep -o '"isLeader":[^,]*' || echo "Failed to get status")
  
  log_info "Manager 1 leadership status: $leader1"
  log_info "Manager 2 leadership status: $leader2"
  
  # Ensure exactly one leader exists
  if [[ "$leader1" == '"isLeader":true' && "$leader2" == '"isLeader":false' ]]; then
    log_info "Manager 1 is currently the leader"
  elif [[ "$leader1" == '"isLeader":false' && "$leader2" == '"isLeader":true' ]]; then
    log_info "Manager 2 is currently the leader"
  else
    log_error "Leader election issue: Either no leader or multiple leaders detected"
  fi
  
  log_success "Manager leader election is working properly"
}

# Test Manager failover
test_manager_failover() {
  log_info "Testing Manager failover..."
  
  # Get initial leader status
  leader1=$(curl -s http://localhost:8081/status | grep -o '"isLeader":[^,]*' || echo "Failed to get status")
  leader2=$(curl -s http://localhost:8082/status | grep -o '"isLeader":[^,]*' || echo "Failed to get status")
  
  log_info "Initial status - Manager 1: $leader1, Manager 2: $leader2"
  
  # Determine which manager to stop (the leader)
  if [[ "$leader1" == '"isLeader":true' ]]; then
    leader_port=8081
    follower_port=8082
    container_to_stop="triggerx-manager-1"
  else
    leader_port=8082
    follower_port=8081
    container_to_stop="triggerx-manager-2"
  fi
  
  log_info "Stopping the current leader ($container_to_stop)..."
  docker stop $container_to_stop
  
  # Wait for failover
  log_info "Waiting for failover to occur..."
  sleep 10
  
  # Check if the follower became the leader
  new_leader_status=$(curl -s http://localhost:$follower_port/status | grep -o '"isLeader":[^,]*' || echo "Failed to get status")
  log_info "New status for remaining manager: $new_leader_status"
  
  if [[ "$new_leader_status" != '"isLeader":true' ]]; then
    # Start the stopped container before reporting error
    docker start $container_to_stop
    sleep 10
    log_error "Failover test failed: The remaining manager did not become the leader"
  fi
  
  # Restart the stopped container
  log_info "Restarting the stopped manager..."
  docker start $container_to_stop
  sleep 10
  
  log_success "Manager failover test passed successfully"
}

# Check and restart manager services if needed
check_manager_services() {
  log_info "Checking manager services..."
  
  # Check Manager 1 service
  log_info "Checking if Manager 1 service is running inside container..."
  if ! docker exec -i triggerx-manager-1 ps aux | grep -q "manager/main.go"; then
    log_warning "Manager 1 service is not running in container. Attempting to restart it..."
    docker exec -i triggerx-manager-1 sh -c "./scripts/services/start-manager.sh &"
    sleep 15
  fi
  
  # Check Manager 2 service
  log_info "Checking if Manager 2 service is running inside container..."
  if ! docker exec -i triggerx-manager-2 ps aux | grep -q "manager/main.go"; then
    log_warning "Manager 2 service is not running in container. Attempting to restart it..."
    docker exec -i triggerx-manager-2 sh -c "./scripts/services/start-manager.sh &"
    sleep 15
  fi
  
  log_success "Manager services checked"
}

# Run all tests
main() {
  echo -e "\n${YELLOW}===============================================${NC}"
  echo -e "${YELLOW}   TriggerX System Test Suite${NC}"
  echo -e "${YELLOW}===============================================${NC}\n"
  
  log_info "Starting comprehensive system test..."
  
  # Check if --start flag is provided
  if [[ "$1" == "--start" ]]; then
    check_docker
    start_containers
    
    # Additional wait time for managers to fully initialize
    log_info "Giving managers extra time to initialize..."
    sleep 30
  else
    check_docker
    check_containers
  fi
  
  # Initialize DB if needed
  if ! docker exec -i triggerx-scylla-1 cqlsh -e "DESCRIBE KEYSPACES" | grep -q "triggerx"; then
    log_warning "triggerx keyspace not found. Initializing database..."
    if [ -f "scripts/database/setup-db.sh" ]; then
      bash scripts/database/setup-db.sh
    else
      log_error "Database initialization script not found at scripts/database/setup-db.sh"
    fi
  fi
  
  # Run the tests
  test_db_cluster
  test_db_replication
  test_db_failover
  
  # Check and restart manager services if needed
  check_manager_services
  
  # Allow more time for managers to fully start if needed
  log_info "Ensuring managers have enough time to initialize fully..."
  sleep 10
  
  test_manager_health
  test_manager_leadership
  test_manager_failover
  
  echo -e "\n${GREEN}===============================================${NC}"
  echo -e "${GREEN}   All tests completed successfully!${NC}"
  echo -e "${GREEN}   The system is working properly.${NC}"
  echo -e "${GREEN}===============================================${NC}\n"
}

# Print usage information
usage() {
  echo "Usage: $0 [--start]"
  echo ""
  echo "Options:"
  echo "  --start    Start required containers if they are not running"
  echo ""
  exit 1
}

# Process command line arguments
if [[ $# -gt 1 ]]; then
  usage
elif [[ $# -eq 1 && "$1" != "--start" ]]; then
  usage
else
  main "$1"
fi