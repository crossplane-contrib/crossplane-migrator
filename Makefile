
DOCKER ?= docker
GO ?= go

PROJECT := crossplane-migrator

GOLANGCI_LINT_VERSION := 1.55.2

build:
	$(GO) build -o $(PROJECT)

lint:
	
	$(DOCKER) run --rm -v $(CURDIR):/app -v ~/.cache/golangci-lint/v$(GOLANGCI_LINT_VERSION):/root/.cache -w /app golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION) golangci-lint run -v
test: 
	$(GO) test ./...

vendor: modules.download 

vendor.check: modules.check

modules.check: modules.tidy.check modules.verify

modules.download:
	$(GO) mod download

modules.tidy:
	$(GO) mod tidy

modules.tidy.check:
	@$(GO) mod tidy
	@changed=$$(git diff --exit-code --name-only go.mod go.sum 2>&1) && [ -z "$${changed}" ] || (echo "go.mod is not tidy. Please run 'make go.modules.tidy' and stage the changes" 1>&2; $(FAIL))

modules.verify:
	$(GO) mod verify

