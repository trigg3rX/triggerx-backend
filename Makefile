############################# HELP MESSAGE #############################
.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

############################# RUN #############################

start-manager: ## Start the task manager
	./scripts/services/start-manager.sh

start-aggregator: ## Start the Aggregator
	./scripts/services/start-aggregator.sh

start-registrar: ## Start the Registrar
	./scripts/services/start-registrar.sh

start-health: ## Start the Health Check
	./scripts/services/start-health.sh

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

############################ KEEPER NODE ####################################

build-keeper: ## Build the Keeper
	./scripts/binary/build.sh

start-keeper: ## Start the Keeper
	./scripts/services/start-keeper.sh
