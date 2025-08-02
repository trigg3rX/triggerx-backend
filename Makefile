############################# HELP MESSAGE #############################
# Make sure the help command stays first, so that it's printed by default when `make` is called without arguments

GO_LINES_IGNORED_DIRS=./othentic/... ./data/... ./docs/... ./scripts/...
GO_PACKAGES=./cmd/... ./internal/... ./pkg/... ./checker/... ./cli/...
GO_FOLDERS=$(shell echo ${GO_PACKAGES} | sed -e "s/\.\///g" | sed -e "s/\/\.\.\.//g")
GO_LINES_IGNORED=$(shell echo ${GO_LINES_IGNORED_DIRS} | sed -e "s/\.\///g" | sed -e "s/\/\.\.\.//g")

.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

############################# DATABASE #############################
----------------------------DATABASE----------------------------: ## 

db-setup: ## Setup ScyllaDB container
	docker compose down
	docker compose up -d
	sleep 6
	./scripts/database/setup-db.sh

db-shell: ## Open CQL shell
	docker exec -it triggerx-scylla cqlsh

db-backup: ## Backup data
	docker exec -it triggerx-scylla nodetool snapshot -t triggerx_backup triggerx -cf keeper_data

############################# RUN #############################
----------------------------SERVICES----------------------------: ## 

start-othentic: ## Start the Othentic Node
	./scripts/services/start-othentic.sh

start-db-server: ## Start the Database Server
	./scripts/services/start-dbserver.sh

start-registrar: ## Start the Registrar
	./scripts/services/start-registrar.sh

start-health: ## Start the Health Check
	./scripts/services/start-health.sh

start-taskmanager: ## Start the Task Manager
	./scripts/services/start-taskmanager.sh

start-time-scheduler: ## Start the Time Scheduler
	./scripts/services/start-time-scheduler.sh

start-condition-scheduler: ## Start the Condition Scheduler
	./scripts/services/start-condition-scheduler.sh

start-keeper: ## Start the Keeper
	./scripts/services/start-keeper.sh

start-imua-keeper: ## Start the Imua Keeper
	./scripts/services/start-imua-keeper.sh

############################ GITHUB ACTIONS ####################################
-------------------------GITHUB-ACTIONS-------------------------: ## 

install-tools: ## Install the tools for GitHub Actions
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
	

format-go: ## Format the Go code (active)
	@which golangci-lint > /dev/null 2>&1 || (echo "Error: golangci-lint is not installed. Please install it first." && exit 1)
	golangci-lint run --fix

dependency-update: ## Update the Go dependencies (inactive)
	@which go-mod-outdated > /dev/null 2>&1 || (echo "Error: go-mod-outdated is not installed. Please install it first." && exit 1)
	go list -u -m -json all | go-mod-outdated -update -direct

build-go: ## Build the Go code (active)
	go build -v ./...
	go mod tidy
	git diff --exit-code go.mod go.sum

############################ BUILD AND PUSH DOCKER IMAGES ####################################
----------------------------DOCKERS----------------------------: ## 

docker-build: ## Build the Docker image (usage: make docker-build SERVICE=<service> VERSION=<version>)
	@if [ -z "$(SERVICE)" ] || [ -z "$(VERSION)" ]; then \
		echo "Error: SERVICE and VERSION are required"; \
		echo "Usage: make docker-build SERVICE=<service> VERSION=<version>"; \
		echo "Example: make docker-build SERVICE=all VERSION=0.0.7"; \
		exit 1; \
	fi
	./scripts/docker/build.sh -n $(SERVICE) -v $(VERSION)

docker-push: ## Push the Docker image
	./scripts/docker/publish.sh

docker-run: ## Run the Docker image, pull if not present
	./scripts/docker/pull-and-run.sh