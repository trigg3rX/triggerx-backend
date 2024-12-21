############################# HELP MESSAGE #############################
.PHONY: help tests
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

############################# CONTRACTS #############################

# build: ## Build the binary
# 	go build -o triggerx
# 	mv triggerx /home/nite-sky/bin/triggerx
# 	# triggerx generate-keystore

eigenlayer: ## Install the EigenLayer CLI and View Instructions
	go install github.com/Layr-Labs/eigenlayer-cli/cmd/eigenlayer@latest
	@echo "EigenLayer CLI installed. View instructions at https://github.com/Layr-Labs/eigenlayer-cli/blob/master/README.md#documentation"

register: ## Register the Keeper to AVS
	go run cli/main.go register