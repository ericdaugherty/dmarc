GOOS?=$(shell go env GOOS)
RACE?=-race

.PHONY: build
build: ## Builds go application to lambda
	pkger
	env GOOS=$(GOOS) go build $(RACE) -o lambda

.PHONY: run
run: build ## Runs the app locally
	./lambda

.PHONY: clean
clean: ## Cleansup
	rm -f lambda

.PHONY: aws
aws: 
	$(eval GOOS=linux)
	$(eval RACE=)

.PHONY: deploy
deploy: clean aws build ## Deploy via Serverless
	sls deploy --verbose

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.DEFAULT_GOAL := help