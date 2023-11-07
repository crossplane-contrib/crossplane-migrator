GO ?= go

PROJECT := crossplane-migrator

build:
	$(GO) build -o $(PROJECT)

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

