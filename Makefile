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

db-cluster-setup: ## Setup ScyllaDB multi-node cluster
	chmod +x ./scripts/database/setup-cluster.sh
	./scripts/database/setup-cluster.sh

db-cluster-status: ## Check ScyllaDB cluster status
	docker exec triggerx-scylla-1 nodetool status

db-test-replication: ## Test ScyllaDB replication and failover
	chmod +x ./scripts/database/test-replication.sh
	./scripts/database/test-replication.sh

db-test-api-failover: ## Test API with node failures
	chmod +x ./scripts/database/test-api-failover.sh
	./scripts/database/test-api-failover.sh

db-api-client: ## Manual API testing client
	chmod +x ./scripts/database/api-test-client.sh
	./scripts/database/api-test-client.sh $(ARGS)

db-shell-node1: ## Open CQL shell for node 1
	docker exec -it triggerx-scylla-1 cqlsh

db-shell-node2: ## Open CQL shell for node 2
	docker exec -it triggerx-scylla-2 cqlsh

start-db-server: ## Start the Database Server
	./scripts/database/start-dbserver.sh

db-shell: ## Open CQL shell
	docker exec -it triggerx-scylla-1 cqlsh

db-backup:  ##backup data
	docker exec -it triggerx-scylla-1 nodetool snapshot -t triggerx_backup triggerx -cf keeper_data

############################# RUN #############################

start-othentic: ## Start the Othentic Node
	./scripts/services/start-othentic.sh

start-manager: ## Start the task manager
	./scripts/services/start-manager.sh

start-registrar: ## Start the Registrar
	./scripts/services/start-registrar.sh

start-health: ## Start the Health Check
	./scripts/services/start-health.sh


############################ KEEPER NODE ####################################

build-keeper: ## Build the Keeper
	./scripts/binary/build.sh

start-keeper: ## Start the Keeper
	./scripts/services/start-keeper.sh
