# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# RSA Keys
# 	To generate a private/public key PEM file.
# 	$ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# 	$ openssl rsa -pubout -in private.pem -out public.pem

##@ Development

run: ## Run the sales service with log formatting
	go run apis/services/sales/main.go | go run apis/tooling/logfmt/main.go

help: ## Display details for all commands
	@awk 'BEGIN {FS = ":.*?##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

version: ## Show application version
	go run apis/services/sales/main.go --version

##@ Testing

curl-test: ## Test the /test endpoint
	curl -il -X GET http://localhost:3000/test

curl-live: ## Test the /liveness endpoint
	curl -il -X GET http://localhost:3000/liveness

curl-ready: ## Test the /readiness endpoint
	curl -il -X GET http://localhost:3000/readiness

curl-error: ## Test the /testerror endpoint
	curl -il -X GET http://localhost:3000/testerror

curl-panic: ## Test the /testpanic endpoint
	curl -il -X GET http://localhost:3000/testpanic

##@ Administration

admin: ## Run the admin tool
	go run apis/tooling/admin/main.go

# admin token
# export TOKEN=eyJhbGciOiJSUzI1NiIsImtpZCI6IjU0YmIyMTY1LTcxZTEtNDFhNi1hZjNlLTdkYTRhMGUxZTJjMSIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzZXJ2aWNlIHByb2plY3QiLCJzdWIiOiIzOGRjOWQ4NC0wMThiLTRhMTUtYjk1OC0wYjc4YWYxMWMzMDEiLCJleHAiOjE3NDU1OTY1MTEsImlhdCI6MTcxNDA2MDUxMSwiUm9sZXMiOlsiQURNSU4iXX0.nioy4PpggnfrwTxNTQKbviCs3duF53Q5jcoRQqdngQSv7lccKYgTmxzuyMano-Yd-KijtHBCZxWAOFEv5w6xGCfqmQRThKXzQXiHN5Cv0OGab5lmThPGRuCHv35lEQzImKU9E1skSwHvCwyX89pRJpnku9PKJMT_Z4oT6amwFA50HU8jSM8j1HQ0ao60jSMgELKFFb9m3u4ZKIj4w7qxwwV9JD2_wH8HWjt5P1L2V5YtnP9vgMOBZ617TTGRysDS8WXQGXqEAiVQVZSneJCUR-4HofXvOTYIKyQG3iUAs3WTf91EubJQbeW6cFmwudE4xx4t20EaVIUkMr00jFxHtg

# user token
# export TOKEN=eyJhbGciOiJSUzI1NiIsImtpZCI6IjU0YmIyMTY1LTcxZTEtNDFhNi1hZjNlLTdkYTRhMGUxZTJjMSIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzZXJ2aWNlIHByb2plY3QiLCJzdWIiOiIzOGRjOWQ4NC0wMThiLTRhMTUtYjk1OC0wYjc4YWYxMWMzMDEiLCJleHAiOjE3NDU1OTY1NTUsImlhdCI6MTcxNDA2MDU1NSwiUm9sZXMiOlsiVVNFUiJdfQ.bBomVD8igAkiXzKvsBQGj5Nb5ho2MYq-eOXeFjtfdauayE4lQamFtWHUjKq1KaIzMvxlybdU_37py620vOBrQ7tUIYTdY91ggf-AgakBTm3UTs494BjIv3rvPM0baXKSMSe8Ao2acjhTSPA9Crz0mEgWynd3gcuSXVAXrSO3eH8bPuBqR2Ohp1_JsE4Y_dj0Mfmzo8wRQHmBI-QVMT0udUlxVGsLhurGmFpLSrQ7Vzr-xuCJPMgjkTZmSvelCrxSeql6-scwZLTHjgqkzqIMS5EceyKAQfYuuRgqrwwIAtGhyM6SrPHH3WFEy2RHW2ebqQKa-fc3JAcJH44hgiiGsA

##@ Authentication

curl-auth: ## Test authenticated endpoint with Bearer token
	curl -il \
	-H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/testauth"

token: ## Get authentication token
	curl -il \
	--user "admin@example.com:gophers" http://localhost:6000/auth/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

users: ## Get users with authentication
	curl -il \
	-H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/users?page=1&rows=2"

curl-auth2: ## Test auth service authentication endpoint
	curl -il \
	-H "Authorization: Bearer ${TOKEN}" "http://localhost:6000/auth/authenticate"

# ==============================================================================
# Deploy First Mentality

# ==============================================================================
# Go Installation
#
#	You need to have Go version 1.24 to run this code.
#
#	https://go.dev/dl/
#
#	If you are not allowed to update your Go frontend, you can install
#	and use a 1.24 frontend.
#
#	$ go install golang.org/dl/go1.24.6@latest
#	$ go1.24.6 download
#
#	This means you need to use `go1.24.6` instead of `go` for any command
#	using the Go frontend tooling from the makefile.

# ==============================================================================
# Brew Installation
#
#	Having brew installed will simplify the process of installing all the tooling.
#
#	Run this command to install brew on your machine. This works for Linux, Mac and Windows.
#	The script explains what it will do and then pauses before it does it.
#	$ /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
#
#	WINDOWS MACHINES
#	These are extra things you will most likely need to do after installing brew
#
# 	Run these three commands in your terminal to add Homebrew to your PATH:
# 	Replace <name> with your username.
#	$ echo '# Set PATH, MANPATH, etc., for Homebrew.' >> /home/<name>/.profile
#	$ echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/<name>/.profile
#	$ eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
#
# 	Install Homebrew's dependencies:
#	$ sudo apt-get install build-essential
#
# 	Install GCC:
#	$ brew install gcc

# ==============================================================================
# Install Tooling and Dependencies
#
#	This project uses Docker and it is expected to be installed. Please provide
#	Docker at least 4 CPUs. To use Podman instead please alias Docker CLI to
#	Podman CLI or symlink the Docker socket to the Podman socket. More
#	information on migrating from Docker to Podman can be found at
#	https://podman-desktop.io/docs/migrating-from-docker.
#
#	Run these commands to install everything needed.
#	$ make dev-brew
#	$ make dev-docker
#	$ make dev-gotooling

# ==============================================================================
# Running Test
#
#	Running the tests is a good way to verify you have installed most of the
#	dependencies properly.
#
#	$ make test

# ==============================================================================
# Running The Project
#
#	$ make dev-up
#	$ make dev-update-apply
#	$ make token
#	$ export TOKEN=<token>
#	$ make users
#
#	You can use `make dev-status` to look at the status of your KIND cluster.

# ==============================================================================
# CLASS NOTES
#
# Kind
# 	For full Kind v0.29 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.29.0
#
# RSA Keys
# 	To generate a private/public key PEM file.
# 	$ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
#	$ openssl rsa -pubout -in private.pem -out public.pem
#	$ ./admin genkey
#
# Testing Coverage
# 	$ go test -coverprofile p.out
# 	$ go tool cover -html p.out
#
# Module Call Examples
# 	$ curl https://proxy.golang.org/github.com/ardanlabs/conf/@v/list
# 	$ curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/list
# 	$ curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.info
# 	$ curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.mod
# 	$ curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.zip
# 	$ curl https://sum.golang.org/lookup/github.com/ardanlabs/conf/v3@v3.1.1
#
# OPA Playground
# 	https://play.openpolicyagent.org/
# 	https://academy.styra.com/
# 	https://www.openpolicyagent.org/docs/latest/policy-reference/

# ==============================================================================
# OrbStack Detection and Setup

# Check if OrbStack is installed and running (macOS only)
ORBSTACK_AVAILABLE := $(shell if [ "$(shell uname)" = "Darwin" ] && command -v orb 2>/dev/null && orb status 2>/dev/null | grep -q "running"; then echo "true"; else echo "false"; fi)
ORBSTACK_K8S_AVAILABLE := $(shell if [ "$(shell uname)" = "Darwin" ] && command -v orb 2>/dev/null && orb k8s status 2>/dev/null | grep -q "enabled"; then echo "true"; else echo "false"; fi)

# Determine if we should use OrbStack or fall back to Kind
USE_ORBSTACK := $(shell if [ "$(ORBSTACK_AVAILABLE)" = "true" ] && [ "$(ORBSTACK_K8S_AVAILABLE)" = "true" ]; then echo "true"; else echo "false"; fi)

# ==============================================================================
# Define dependencies

GOLANG          := golang:1.24
ALPINE          := alpine:3.19
KIND            := kindest/node:v1.29.2
POSTGRES        := postgres:16.2
GRAFANA         := grafana/grafana:10.4.0
PROMETHEUS      := prom/prometheus:v2.51.0
TEMPO           := grafana/tempo:2.4.0
LOKI            := grafana/loki:2.9.0
PROMTAIL        := grafana/promtail:2.9.0

KIND_CLUSTER    := ardan-starter-cluster
NAMESPACE       := sales-system
SALES_APP       := sales
AUTH_APP        := auth
BASE_IMAGE_NAME := localhost/ardanlabs
VERSION         := 0.0.1
SALES_IMAGE     := $(BASE_IMAGE_NAME)/$(SALES_APP):$(VERSION)
METRICS_IMAGE   := $(BASE_IMAGE_NAME)/metrics:$(VERSION)
AUTH_IMAGE      := $(BASE_IMAGE_NAME)/$(AUTH_APP):$(VERSION)

##@ Docker

build: sales auth ## Build all Docker images

sales: ## Build sales service Docker image
	docker build \
		-f zarf/docker/dockerfile.sales \
		-t $(SALES_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		.

auth: ## Build auth service Docker image
	docker build \
		-f zarf/docker/dockerfile.auth \
		-t $(AUTH_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
		.

##@ Kubernetes

orbstack-install: ## Install OrbStack (macOS only)
	@echo "Installing OrbStack..."
	@if [ "$(shell uname)" = "Darwin" ]; then \
		curl -fsSL https://update.orbstack.dev/install.sh | sh; \
		echo "OrbStack installed. Please restart your terminal and run 'orb start'"; \
		echo "Then run 'make orbstack-setup' to enable Kubernetes"; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		echo "OrbStack is not available on Linux. Using Kind instead."; \
		echo "Run 'make dev-up' to use Kind cluster."; \
	elif [ "$(shell uname)" = "MINGW64_NT" ] || [ "$(shell uname)" = "MSYS_NT" ]; then \
		echo "OrbStack is not available on Windows. Using Kind instead."; \
		echo "Run 'make dev-up' to use Kind cluster."; \
	else \
		echo "OrbStack is only available on macOS. Using Kind instead."; \
		echo "Run 'make dev-up' to use Kind cluster."; \
	fi

orbstack-setup: ## Setup OrbStack Kubernetes
	@echo "Setting up OrbStack Kubernetes..."
	@if [ "$(ORBSTACK_AVAILABLE)" = "true" ]; then \
		orb k8s enable; \
		echo "OrbStack Kubernetes enabled. Run 'orb k8s status' to verify."; \
	else \
		echo "OrbStack is not running. Please run 'orb start' first."; \
		exit 1; \
	fi

dev-up: ## Create and start Kubernetes cluster (OrbStack or Kind)
	@if [ "$(USE_ORBSTACK)" = "true" ]; then \
		echo "Using OrbStack Kubernetes..."; \
		kubectl cluster-info; \
	elif [ "$(ORBSTACK_AVAILABLE)" = "true" ] && [ "$(ORBSTACK_K8S_AVAILABLE)" = "false" ]; then \
		echo "OrbStack is running but Kubernetes is not enabled."; \
		echo "Run 'make orbstack-setup' to enable Kubernetes."; \
		exit 1; \
	elif [ "$(ORBSTACK_AVAILABLE)" = "false" ]; then \
		echo "OrbStack not detected. Installing and using Kind instead..."; \
		echo "To use OrbStack instead, run 'make orbstack-install' then 'make orbstack-setup'"; \
		kind create cluster \
			--image $(KIND) \
			--name $(KIND_CLUSTER) \
			--config zarf/k8s/dev/kind-config.yaml; \
		kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner; \
		kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER); \
	else \
		echo "OrbStack is not running. Please run 'orb start' first."; \
		exit 1; \
	fi

dev-down: ## Delete Kubernetes cluster
	@if [ "$(USE_ORBSTACK)" = "true" ]; then \
		echo "OrbStack Kubernetes cluster cannot be deleted separately."; \
		echo "Use 'orb k8s disable' to disable Kubernetes if needed."; \
	else \
		kind delete cluster --name $(KIND_CLUSTER); \
	fi

dev-status-all: ## Show all Kubernetes resources
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-status: ## Watch pod status
	watch -n 2 kubectl get pods -o wide --all-namespaces

# ------------------------------------------------------------------------------

dev-load-db: ## Load PostgreSQL image into cluster
	@if [ "$(USE_ORBSTACK)" = "true" ]; then \
		echo "OrbStack automatically has access to local Docker images"; \
	else \
		kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER); \
	fi

dev-load: ## Load application images into cluster
	@if [ "$(USE_ORBSTACK)" = "true" ]; then \
		echo "OrbStack automatically has access to local Docker images"; \
	else \
		kind load docker-image $(SALES_IMAGE) --name $(KIND_CLUSTER); \
		kind load docker-image $(AUTH_IMAGE) --name $(KIND_CLUSTER); \
	fi

dev-apply: ## Apply all Kubernetes manifests
	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database

	kustomize build zarf/k8s/dev/auth | kubectl apply -f -
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(AUTH_APP) --timeout=120s --for=condition=Ready

	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(SALES_APP) --timeout=120s --for=condition=Ready

dev-restart: ## Restart all deployments
	kubectl rollout restart deployment $(AUTH_APP) --namespace=$(NAMESPACE)
	kubectl rollout restart deployment $(SALES_APP) --namespace=$(NAMESPACE)

dev-update: build dev-load dev-restart ## Build, load, and restart deployments

dev-update-apply: build dev-load dev-apply ## Build, load, and apply all manifests

dev-logs: ## Follow sales service logs
	kubectl logs --namespace=$(NAMESPACE) -l app=$(SALES_APP) --all-containers=true -f --tail=100 --max-log-requests=6 | go run apis/tooling/logfmt/main.go -service=$(SALES_APP)

dev-logs-auth: ## Follow auth service logs
	kubectl logs --namespace=$(NAMESPACE) -l app=$(AUTH_APP) --all-containers=true -f --tail=100 | go run apis/tooling/logfmt/main.go

dev-logs-init: ## Follow init container logs
	kubectl logs --namespace=$(NAMESPACE) -l app=$(SALES_APP) -f --tail=100 -c init-migrate-seed

# ------------------------------------------------------------------------------

dev-describe-deployment: ## Describe sales deployment
	kubectl describe deployment --namespace=$(NAMESPACE) $(SALES_APP)

dev-describe-sales: ## Describe sales pods
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(SALES_APP)

dev-describe-auth: ## Describe auth pods
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(AUTH_APP)

##@ Monitoring

metrics: ## Monitor application metrics with expvarmon
	expvarmon -ports="localhost:3010" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

statsviz: ## Open stats visualization in browser
	@if [ "$(shell uname)" = "Darwin" ]; then \
		open http://localhost:3010/debug/statsviz; \
	elif [ "$(shell uname)" = "Linux" ]; then \
		if command -v xdg-open >/dev/null 2>&1; then \
			xdg-open http://localhost:3010/debug/statsviz; \
		elif command -v sensible-browser >/dev/null 2>&1; then \
			sensible-browser http://localhost:3010/debug/statsviz; \
		else \
			echo "Please open http://localhost:3010/debug/statsviz in your browser"; \
		fi; \
	elif [ "$(shell uname)" = "MINGW64_NT" ] || [ "$(shell uname)" = "MSYS_NT" ]; then \
		start http://localhost:3010/debug/statsviz; \
	else \
		echo "Please open http://localhost:3010/debug/statsviz in your browser"; \
	fi

##@ Infrastructure

terraform-init: ## Initialize Terraform
	@echo "Initializing Terraform..."
	cd terraform && terraform init

terraform-plan: ## Plan Terraform changes
	@echo "Planning Terraform changes..."
	cd terraform && terraform plan

terraform-apply: ## Apply Terraform changes
	@echo "Applying Terraform changes..."
	cd terraform && terraform apply -auto-approve

terraform-destroy: ## Destroy Terraform infrastructure
	@echo "⚠️  WARNING: This will destroy all infrastructure!"
	@read -p "Are you sure? Type 'yes' to confirm: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		cd terraform && terraform destroy -auto-approve; \
	else \
		echo "Destroy cancelled."; \
	fi

terraform-output: ## Show Terraform outputs
	@echo "Terraform outputs:"
	cd terraform && terraform output

terraform-secrets: ## Generate GitHub secrets from Terraform outputs
	@echo "Generating GitHub secrets values..."
	cd terraform && ./get-secrets.sh

terraform-refresh: ## Refresh Terraform state
	@echo "Refreshing Terraform state..."
	cd terraform && terraform refresh

terraform-validate: ## Validate Terraform configuration
	@echo "Validating Terraform configuration..."
	cd terraform && terraform validate

terraform-fmt: ## Format Terraform files
	@echo "Formatting Terraform files..."
	cd terraform && terraform fmt -recursive

terraform-clean: ## Clean Terraform files and state
	@echo "Cleaning Terraform files..."
	cd terraform && rm -rf .terraform .terraform.lock.hcl terraform.tfstate*

# Convenience targets
terraform-setup: terraform-init terraform-validate terraform-plan ## Initialize, validate, and plan Terraform

terraform-deploy: terraform-init terraform-validate terraform-apply ## Deploy infrastructure (init, validate, apply)

terraform-redeploy: terraform-destroy terraform-apply ## Destroy and redeploy infrastructure

terraform-status: terraform-output terraform-secrets ## Show infrastructure status and secrets

##@ Database

pgcli: ## Connect to PostgreSQL database
	pgcli postgresql://postgres:postgres@localhost

##@ Dependencies

dev-brew: ## Install development tools via Homebrew
	@echo "Checking Homebrew installation..."
	@if ! command -v brew >/dev/null 2>&1; then \
		echo "Homebrew not found. Please install Homebrew first:"; \
		echo "  /bin/bash -c \"\$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""; \
		exit 1; \
	fi
	@echo "Updating Homebrew..."
	@brew update
	@echo ""
	@echo "Checking Go installation..."
	@if command -v go >/dev/null 2>&1; then \
		echo "Go is already installed: $$(go version)"; \
		echo "Available packages: go@1.24, kind, kubectl, kustomize, pgcli, watch, terraform"; \
	else \
		echo "Go is not installed. Available packages: go@1.24, kind, kubectl, kustomize, pgcli, watch, terraform"; \
	fi
	@echo "Enter the packages you want to install (space-separated, or 'all' for all):"
	@read -p "Packages to install: " packages; \
	if [ "$$packages" = "all" ]; then \
		if ! command -v go >/dev/null 2>&1; then brew list go@1.24 || brew install go@1.24; fi; \
		brew list kind || brew install kind; \
		brew list kubectl || brew install kubectl; \
		brew list kustomize || brew install kustomize; \
		brew list pgcli || brew install pgcli; \
		brew list watch || brew install watch; \
		brew list terraform || brew install terraform; \
		echo "All packages installed successfully."; \
	else \
		for pkg in $$packages; do \
			if [ "$$pkg" = "go@1.24" ] && ! command -v go >/dev/null 2>&1; then brew list go@1.24 || brew install go@1.24; fi; \
			if [ "$$pkg" = "kind" ]; then brew list kind || brew install kind; fi; \
			if [ "$$pkg" = "kubectl" ]; then brew list kubectl || brew install kubectl; fi; \
			if [ "$$pkg" = "kustomize" ]; then brew list kustomize || brew install kustomize; fi; \
			if [ "$$pkg" = "pgcli" ]; then brew list pgcli || brew install pgcli; fi; \
			if [ "$$pkg" = "watch" ]; then brew list watch || brew install watch; fi; \
			if [ "$$pkg" = "terraform" ]; then brew list terraform || brew install terraform; fi; \
		done; \
		echo "Selected packages installed successfully."; \
	fi

dev-docker: ## Pull required Docker images
	@echo "Pulling required Docker images..."
	docker pull $(GOLANG)
	docker pull $(ALPINE)
	docker pull $(KIND)
	docker pull $(POSTGRES)
	docker pull $(GRAFANA)
	docker pull $(PROMETHEUS)
	docker pull $(TEMPO)
	docker pull $(LOKI)
	docker pull $(PROMTAIL)
	@echo "All Docker images pulled successfully."

dev-gotooling: ## Install Go development tools
	@echo "Installing Go development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/divan/expvarmon@latest
	@echo "Go development tools installed successfully."

dev-terraform: ## Install Terraform
	@echo "Installing Terraform..."
	@if command -v terraform >/dev/null 2>&1; then \
		echo "Terraform is already installed: $$(terraform version)"; \
	else \
		if command -v brew >/dev/null 2>&1; then \
			brew install terraform; \
		else \
			echo "Homebrew not found. Please install Terraform manually:"; \
			echo "  https://developer.hashicorp.com/terraform/downloads"; \
			exit 1; \
		fi; \
	fi
	@echo "Terraform installation complete."

tidy: ## Tidy and vendor Go modules
	go mod tidy
	go mod vendor

##@ Testing

test-r: ## Run tests with race detection
	CGO_ENABLED=1 go test -race -count=1 ./...

test-only: ## Run tests without race detection
	CGO_ENABLED=0 go test -count=1 ./...

lint: ## Run linting checks
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...

vuln-check: ## Check for vulnerabilities
	govulncheck ./...

test: test-only lint ## Run all tests and checks

test-race: test-r lint ## Run tests with race detection and all checks

test-full: test vuln-check ## Run all tests including vulnerability checks

##@ Administration

migrate: ## Run database migrations
	export SALES_DB_HOST=localhost; go run apis/tooling/admin/main.go migrate

seed: migrate ## Run database migrations and seed data
	export SALES_DB_HOST=localhost; go run apis/tooling/admin/main.go seed

token-gen: ## Generate authentication token
	export SALES_DB_HOST=localhost; go run apis/tooling/admin/main.go gentoken 5cf37266-3473-4006-984f-9325122678b7 54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

##@ Docker Compose

compose-up: ## Start services with Docker Compose
	cd ./zarf/compose/ && docker compose -f docker_compose.yaml -p compose up -d

compose-build-up: build compose-up ## Build and start services with Docker Compose

compose-down: ## Stop Docker Compose services
	cd ./zarf/compose/ && docker compose -f docker_compose.yaml down

compose-logs: ## Show Docker Compose logs
	cd ./zarf/compose/ && docker compose -f docker_compose.yaml logs

##@ Admin Frontend

ADMIN_FRONTEND_PREFIX := ./api/frontends/admin

write-token-to-env: ## Write token to admin frontend environment
	echo "VITE_SERVICE_API=http://localhost:3000/v1" > ${ADMIN_FRONTEND_PREFIX}/.env
	make token | grep -o '"ey.*"' | awk '{print "VITE_SERVICE_TOKEN="$$1}' >> ${ADMIN_FRONTEND_PREFIX}/.env

admin-gui-install: ## Install admin frontend dependencies
	pnpm -C ${ADMIN_FRONTEND_PREFIX} install

admin-gui-update: ## Update admin frontend dependencies
	pnpm -C ${ADMIN_FRONTEND_PREFIX} update

admin-gui-dev: admin-gui-install ## Start admin frontend in development mode
	pnpm -C ${ADMIN_FRONTEND_PREFIX} run dev

admin-gui-build: admin-gui-install ## Build admin frontend
	pnpm -C ${ADMIN_FRONTEND_PREFIX} run build

admin-gui-start-build: admin-gui-build ## Start built admin frontend
	pnpm -C ${ADMIN_FRONTEND_PREFIX} run preview

admin-gui-run: write-token-to-env admin-gui-start-build ## Run admin frontend with token

##@ Load Testing

load: ## Run load test with hey
	hey -m GET -c 100 -n 1000 \
	-H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/v1/users?page=1&rows=2"

load-hack: ## Run load test on hack endpoint
	hey -m GET -c 100 -n 100000 "http://localhost:3000/v1/hack"

##@ Module Management

deps-reset: ## Reset dependencies to original state
	git checkout -- go.mod
	go mod tidy
	go mod vendor

deps-list: ## List all dependencies
	go list -m -u -mod=readonly all

deps-upgrade: ## Upgrade all dependencies
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-cleancache: ## Clean Go module cache
	go clean -modcache

list: ## List all modules
	go list -mod=mod all
