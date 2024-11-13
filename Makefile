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