############################# HELP MESSAGE #############################
.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

############################# DATABASE #############################

db-setup: ## Setup ScyllaDB container
	docker compose down
	docker compose up -d
	sleep 6
	./scripts/database/setup-db.sh

start-db-server: ## Start the Database Server
	./scripts/database/start-dbserver.sh

db-shell: ## Open CQL shell
	docker exec -it triggerx-scylla cqlsh

db-backup:  ##backup data
	docker exec -it triggerx-scylla nodetool snapshot -t triggerx_backup triggerx -cf keeper_data

############################# RUN #############################

start-othentic: ## Start the Othentic Node
	./scripts/services/start-othentic.sh

start-registrar: ## Start the Registrar
	./scripts/services/start-registrar.sh

start-health: ## Start the Health Check
	./scripts/services/start-health.sh

start-redis: ## Start the Redis
	./scripts/services/start-redis.sh

start-time-scheduler: ## Start the Time Scheduler
	./scripts/services/start-time-scheduler.sh

start-event-schedulers: ## Start the Event Schedulers
	./scripts/services/start-event-schedulers.sh

start-condition-scheduler: ## Start the Condition Scheduler
	./scripts/services/start-condition-scheduler.sh

############################ KEEPER NODE ####################################

build-keeper: ## Build the Keeper
	./scripts/binary/build.sh

start-keeper: ## Start the Keeper
	./scripts/services/start-keeper.sh

############################ GITHUB ACTIONS ####################################

install-tools-for-github-actions: ## Install the tools for GitHub Actions
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/psampaz/go-mod-outdated@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

format-go: ## Format the Go code
	@which golangci-lint > /dev/null 2>&1 || (echo "Error: golangci-lint is not installed. Please install it first." && exit 1)
	golangci-lint run --fix

dependency-update: ## Update the Go dependencies
	@which go-mod-outdated > /dev/null 2>&1 || (echo "Error: go-mod-outdated is not installed. Please install it first." && exit 1)
	go list -u -m -json all | go-mod-outdated -update -direct

build-go: ## Build the Go code
	go build -v ./...
	go mod tidy
	git diff --exit-code go.mod go.sum

security-scan: ## Scan the Go code for security vulnerabilities
	@which gosec > /dev/null 2>&1 || (echo "Error: gosec is not installed. Please install it first." && exit 1)
	gosec ./...
