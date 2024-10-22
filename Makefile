############################# HELP MESSAGE #############################
.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

############################# CONTRACTS #############################

start-taskmanager: ## Start the taskmanager
	./utils/start-taskmanager.sh

start-aggregator: ## Start the aggregator
	./utils/start-aggregator.sh

start-keeper: ## Start the keeper 
	./utils/start-keeper.sh

start-docker: ## Start the Keeper in Docker
	./utils/start-docker.sh