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
db-setup:
    docker compose down
    docker compose up -d
    sleep 6
    ./scripts/database/setup-db.sh

# Open CQL shell
db-shell:
    docker exec -it triggerx-scylla cqlsh

# Backup data
db-backup:
    docker exec -it triggerx-scylla nodetool snapshot -t triggerx_backup triggerx -cf keeper_data

############################# SERVICES #############################

# Start the Othentic Node
start-othentic:
    ./scripts/services/start-othentic.sh

# Start the Database Server
start-db-server:
    ./scripts/services/start-dbserver.sh

# Start the Health Check
start-health:
    ./scripts/services/start-health.sh

# Start the Task Dispatcher
start-taskdispatcher:
    ./scripts/services/start-taskdispatcher.sh

# Start the Task Monitor
start-taskmonitor:
    ./scripts/services/start-taskmonitor.sh

# Start the Time Scheduler
start-time-scheduler:
    ./scripts/services/start-time-scheduler.sh

# Start the Condition Scheduler
start-condition-scheduler:
    ./scripts/services/start-condition-scheduler.sh

# Start the Keeper
start-keeper:
    ./scripts/services/start-keeper.sh

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
    go install github.com/golang/mock/mockgen@latest
    go install github.com/axw/gocov/gocov@latest
    go install github.com/matm/gocov-html/cmd/gocov-html@latest
# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6

# Format the Go code (active)
format-go:
    @which golangci-lint > /dev/null 2>&1 || (echo "Error: golangci-lint is not installed. Please install it first." && exit 1)
    golangci-lint run --fix

# Update the Go dependencies (inactive)
dependency-update:
    @which go-mod-outdated > /dev/null 2>&1 || (echo "Error: go-mod-outdated is not installed. Please install it first." && exit 1)
    go list -u -m -json all | go-mod-outdated -update -direct

# Build the Go code (active)
build-go:
    go build -v ./...
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
docker-push:
    ./scripts/docker/publish.sh

# Run the Docker image, pull if not present
docker-run:
    ./scripts/docker/pull-and-run.sh
