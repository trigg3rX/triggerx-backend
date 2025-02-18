############################# HELP MESSAGE #############################
.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


############################# BINARY #############################

build-binary: ## Build the binary
	./scripts/binary/build.sh

release-binary: ## Build the binary for release
	@if [ -z "$(version)" ]; then \
		echo "Error: version argument is required"; \
		echo "Usage: make release-binary version=<version>"; \
		exit 1; \
	fi
	./scripts/binary/release.sh $(version)

############################# RUN #############################

start-manager: ## Start the task manager
	./scripts/start-manager.sh

start-validator: ## Start the task validator
	./scripts/start-validator.sh

start-quorum: ## Start the quorum Manager
	./scripts/start-quorum.sh

start-keeper: ## Start the keeper node
	./scripts/start-keeper.sh

############################# DATABASE #############################

db-setup: ## Setup ScyllaDB container
	./scripts/database/setup-db.sh

start-db-server: ## Start the Database Server
	./scripts/database/start-dbserver.sh

db-shell: ## Open CQL shell
	docker exec -it triggerx-scylla cqlsh

############################ KEEPER NODE ####################################

run-keeper: ## Build the keeper node
	./scripts/build-keeper-node.sh

publish-keeper-node: ## Publish the keeper node
	./scripts/publish-keeper-node.sh

