GOOS?=$(shell go env GOOS)

.PHONY: build
build: ## Builds go application
	env GOOS=$(GOOS) go build -o inbound

.PHONY: deploy
deploy: clean aws build ## Deploy to AWS via Serverless
	sls deploy

.PHONY: clean
clean: ## Delete output files.
	rm -f inbound

.PHONY: log
log: ## Tail Lambda logfile
	serverless logs -f inbound -t

.PHONY: aws
aws: 
	$(eval GOOS=linux)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.DEFAULT_GOAL := help