# TriggerX Backend Justfile
# A modern task runner for the TriggerX backend project

# Variables
go-lines-ignored-dirs := "./othentic/... ./data/... ./docs/... ./scripts/..."
go-packages := "./cmd/... ./internal/... ./pkg/... ./checker/... ./cli/..."

# Default recipe - show help
default:
    @just --list

# Help message
help:
    @just --list

############################# DATABASE #############################

# Setup ScyllaDB container
# Uses the same database environment variables as the application config:
# DATABASE_USERNAME, DATABASE_PASSWORD (from .env file)
# See: internal/dbserver/config/config.go and pkg/env/common.go
db-setup:
    docker compose -f docker/docker-compose.yaml --profile scylla down
    docker compose -f docker/docker-compose.yaml --profile scylla up -d
    sleep 6
    ./scripts/database/setup-db.sh

# Open CQL shell
db-shell:
    docker exec -it triggerx-scylla cqlsh

# Backup data
db-backup:
    docker exec -it triggerx-scylla nodetool snapshot -t triggerx_backup triggerx -cf keeper_data

########################### OBSERVABILITY #########################

# Start the Observability Services
start-observability:
    ./scripts/observability/start-observability.sh

# Stop the Observability Services
stop-observability:
    ./scripts/observability/stop-observability.sh

############################# SERVICES #############################

# Start the Othentic Node
start-othentic:
    ./scripts/services/start-othentic.sh

# Start the Database Server
start-db-server args="":
    ./scripts/services/start-dbserver.sh {{args}}

# Start the Health Check
start-health args="":
    ./scripts/services/start-health.sh {{args}}

# Start the Task Dispatcher
start-taskdispatcher args="":
    ./scripts/services/start-taskdispatcher.sh {{args}}

# Start the Task Monitor
start-taskmonitor args="":
    ./scripts/services/start-taskmonitor.sh {{args}}

# Start the Time Scheduler
start-time-scheduler args="":
    ./scripts/services/start-time-scheduler.sh {{args}}

# Start the Condition Scheduler
start-condition-scheduler args="":
    ./scripts/services/start-condition-scheduler.sh {{args}}

# Start the Event Monitor
start-eventmonitor args="":
    ./scripts/services/start-eventmonitor.sh {{args}}

# Start the Keeper
start-keeper args="":
    ./scripts/services/start-keeper.sh {{args}}

# Start the Imua Keeper
start-imua-keeper:
    ./scripts/services/start-imua-keeper.sh

############################# TESTING #############################

# Run all tests
test:
    go test -v ./...

# Run tests in short mode
test-short:
    go test -v -short ./...

# Run tests with race detection
test-race:
    go test -v -race ./...

# Run tests with coverage
test-coverage:
    @echo "Use the cmd: ./scripts/tests/coverage.sh <FOLDER_PATH>"
    @echo "Example: just test-coverage ./internal/keeper/..."
    @echo "Note: If no folder path is provided, it will run all tests"

# Run unit tests only
test-unit:
    go test -v -short ./internal/... ./pkg/...

# Run integration tests only
test-integration:
    go test -v -run Integration ./...

# Run API tests only
test-api:
    go test -v ./internal/*/api/...

# Run database tests only
test-database:
    go test -v ./internal/*/repository/... ./pkg/database/...

# Run benchmarks
benchmark:
    go test -v -bench=. -benchmem ./pkg/...

############################ GITHUB ACTIONS ####################################

# Install the tools for GitHub Actions
install-tools:
    ./scripts/go/install-tools.sh

# Format the Go code (active)
format-go:
    @which golangci-lint > /dev/null 2>&1 || (echo "Error: golangci-lint is not installed. Please install it first using install-tools." && exit 1)
    golangci-lint run --fix

# Build the Go code (active)
build-go:
    go build -v ./cmd/... ./internal/... ./pkg/...
    go mod tidy
    git diff --exit-code go.mod go.sum

############################ BUILD AND PUSH DOCKER IMAGES ####################################

# Build the Docker image
# Usage: just docker-build <service> <version>
# Example: just docker-build all 0.0.7
docker-build service version:
    @if [ -z "{{service}}" ] || [ -z "{{version}}" ]; then \
        echo "Error: SERVICE and VERSION are required"; \
        echo "Usage: just docker-build <service> <version>"; \
        echo "Example: just docker-build all 0.0.7"; \
        exit 1; \
    fi
    ./scripts/docker/build.sh -n {{service}} -v {{version}}

# Push the Docker image
docker-push service version:
    @if [ -z "{{service}}" ] || [ -z "{{version}}" ]; then \
        echo "Error: SERVICE and VERSION are required"; \
        echo "Usage: just docker-push <service> <version>"; \
        echo "Example: just docker-push all 0.0.7"; \
        exit 1; \
    fi
    ./scripts/docker/publish.sh -n {{service}} -v {{version}}

# Run the Docker image, pull if not present
docker-run:
    ./scripts/docker/pull-and-run.sh
