# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	@echo 'Removing test cache...'
	go clean -testcache
	@echo 'Running tests...'
	go test -race -vet=off -timeout 15s ./...

## bench: run all benchmarks
.PHONY: bench
bench:
	@echo 'Running benchmarks...'
	go test -bench=.

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: tidy
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Linting code...'
	golangci-lint run
	@echo 'Running tests...'
	go test -race -vet=off ./...

## tidy: tidy and verify dependencies
.PHONY: tidy
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
