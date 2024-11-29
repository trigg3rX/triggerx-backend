############################# HELP MESSAGE #############################
.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

############################# RUN #############################

start-manager: ## Start the task manager
	./scripts/start-manager.sh

start-validator: ## Start the task validator
	./scripts/start-validator.sh

start-quorumcreator: ## Start the quorum creator
	./scripts/start-quorumcreator.sh


############################# DATABASE #############################

db-start: ## Start ScyllaDB container
	docker-compose up -d

db-init: ## Initialize database schema
	./scripts/init-db.sh

db-shell: ## Open CQL shell
	docker exec -it triggerx-scylla cqlsh

############################# FULL SETUP #############################

setup: db-start db-init ## Setup everything

start-api: ## Start the API server
	./scripts/start-api.sh

############################# GENERATE BINDINGS #############################

generate-bindings: ## Generate bindings
	./pkg/avsinterface/generate-bindings.sh
