SHELL := /bin/bash

# Deploy First Mentality

# ==============================================================================
# Brew Installation
#
#	Having brew installed will simplify the process of installing all the tooling.
#
#	Run this command to install brew on your machine. This works for Linux, Max and Windows.
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
# Windows Users ONLY - Install Telepresence
#
#	Unfortunately you can't use brew to install telepresence because you will
#	receive a bad binary. Please follow these instruction.
#
#	$ sudo curl -fL https://app.getambassador.io/download/tel2/linux/amd64/latest/telepresence -o /usr/local/bin/telepresence
#	$ sudo chmod a+x /usr/local/bin/telepresence
#
# 	Restart your wsl environment.

# ==============================================================================
# Install Tooling and Dependencies 
#
#	If you are running a mac or linux machine with brew, run these commands:
#	$ make dev-brew
#	$ make dev-docker
#	$ make dev-gotooling
#
#	If you are a windows user and have installed brew, run these commands:
#	$ make dev-brew-common
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
# Starting The Project
#
#	If you want to use telepresence (recommended):
#	$ make dev-up
#	$ make dev-update-apply 
#
#	If telepresence is not working for you:
#	$ make dev-up-local
#	$ make dev-update-apply
#
#	Note: If you attempted to run with telepresence and it didn't work, you may
#		  want to restart the cluser.
#		  $ make dev-down-local
#
#	Note: When running without telepresence, if you see a command where there is
#         a `-local` option, you will need to use that command.

# ==============================================================================
# CLASS NOTES
#
# Kind
# 	For full Kind v0.18 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.18.0
#
# RSA Keys
# 	To generate a private/public key PEM file.
# 	$ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# 	$ openssl rsa -pubout -in private.pem -out public.pem
# 	$ ./sales-admin genkey
#
# Testing Coverage
# 	$ go test -coverprofile p.out
# 	$ go tool cover -html p.out
#
# Hashicorp Vault
# 	READ THIS: https://developer.hashicorp.com/vault/docs/concepts/tokens
# 	$ export VAULT_TOKEN=mytoken
# 	$ export VAULT_ADDR='http://vault-service.sales-system.svc.cluster.local:8200'
# 	$ vault secrets list
# 	$ vault kv get secret/sales
# 	$ vault kv put secret/sales key="some data"
# 	$ kubectl logs --namespace=sales-system -l app=sales -c init-vault-server
# 	$ curl -H "X-Vault-Token: mytoken" -X GET http://vault-service.sales-system.svc.cluster.local:8200/v1/secret/data/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
# 	$ curl -H "X-Vault-Token: mytoken" -H "Content-Type: application/json" -X POST -d '{"data":{"pk":"PEM"}}' http://vault-service.sales-system.svc.cluster.local:8200/v1/secret/data/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
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
# Install dependencies

GOLANG       := golang:1.20
ALPINE       := alpine:3.17
KIND         := kindest/node:v1.26.3
POSTGRES     := postgres:15-alpine
VAULT        := hashicorp/vault:1.13
ZIPKIN       := openzipkin/zipkin:2.24
TELEPRESENCE := docker.io/datawire/tel2:2.12.2

dev-brew-common:
	brew update
	brew tap hashicorp/tap
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize
	brew list pgcli || brew install pgcli
	brew list vault || brew install vault

dev-brew: dev-brew-common
	brew list datawire/blackbird/telepresence || brew install datawire/blackbird/telepresence

dev-brew-arm64: dev-brew-common
	brew list datawire/blackbird/telepresence-arm64 || brew install datawire/blackbird/telepresence-arm64

dev-docker:
	docker pull $(GOLANG)
	docker pull $(ALPINE)
	docker pull $(KIND)
	docker pull $(POSTGRES)
	docker pull $(VAULT)
	docker pull $(ZIPKIN)
	docker pull $(TELEPRESENCE)

dev-gotooling:
	go install github.com/divan/expvarmon@latest
	go install github.com/rakyll/hey@latest

# ==============================================================================
# Building containers

# Example: $(shell git rev-parse --short HEAD)
VERSION := 1.0

all: sales metrics

sales:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

metrics:
	docker build \
		-f zarf/docker/dockerfile.metrics \
		-t metrics:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/kind

KIND_CLUSTER := ardan-starter-cluster

dev-up-local:
	kind create cluster \
		--image kindest/node:v1.26.3@sha256:61b92f38dff6ccc29969e7aa154d34e38b89443af1a2c14e6cfbd2df6419c66f \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml
	
	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)
	kind load docker-image $(VAULT) --name $(KIND_CLUSTER)
	kind load docker-image $(ZIPKIN) --name $(KIND_CLUSTER)

dev-up: dev-up-local
	telepresence --context=kind-$(KIND_CLUSTER) helm install
	telepresence --context=kind-$(KIND_CLUSTER) connect

dev-down-local:
	kind delete cluster --name $(KIND_CLUSTER)

dev-down:
	telepresence quit -s
	kind delete cluster --name $(KIND_CLUSTER)

# ------------------------------------------------------------------------------

dev-load:
	cd zarf/k8s/dev/sales; kustomize edit set image sales-api-image=sales-api:$(VERSION)
	kind load docker-image sales-api:$(VERSION) --name $(KIND_CLUSTER)

	cd zarf/k8s/dev/sales; kustomize edit set image metrics-image=metrics:$(VERSION)
	kind load docker-image metrics:$(VERSION) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build zarf/k8s/dev/vault | kubectl apply -f -

	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl wait pods --namespace=sales-system --selector app=database --for=condition=Ready
	
	kustomize build zarf/k8s/dev/zipkin | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/zipkin
	
	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/sales

dev-restart:
	kubectl rollout restart deployment sales --namespace=sales-system

dev-update: all dev-load dev-restart

dev-update-apply: all dev-load dev-apply

# ------------------------------------------------------------------------------

dev-logs:
	kubectl logs --namespace=sales-system -l app=sales --all-containers=true -f --tail=100 --max-log-requests=6 | go run app/tooling/logfmt/main.go -service=SALES-API

dev-logs-init:
	kubectl logs --namespace=sales-system -l app=sales -f --tail=100 -c init-vault-system
	kubectl logs --namespace=sales-system -l app=sales -f --tail=100 -c init-vault-loadkeys
	kubectl logs --namespace=sales-system -l app=sales -f --tail=100 -c init-migrate
	kubectl logs --namespace=sales-system -l app=sales -f --tail=100 -c init-seed

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-describe:
	kubectl describe nodes
	kubectl describe svc

dev-describe-deployment:
	kubectl describe deployment --namespace=sales-system sales

dev-describe-sales:
	kubectl describe pod --namespace=sales-system -l app=sales

# ------------------------------------------------------------------------------

dev-logs-vault:
	kubectl logs --namespace=sales-system -l app=vault --all-containers=true -f --tail=100

dev-logs-db:
	kubectl logs --namespace=sales-system -l app=database --all-containers=true -f --tail=100

dev-logs-zipkin:
	kubectl logs --namespace=sales-system -l app=zipkin --all-containers=true -f --tail=100

# ------------------------------------------------------------------------------

dev-services-delete:
	kustomize build zarf/k8s/dev/sales | kubectl delete -f -
	kustomize build zarf/k8s/dev/zipkin | kubectl delete -f -
	kustomize build zarf/k8s/dev/database | kubectl delete -f -

dev-describe-replicaset:
	kubectl get rs
	kubectl describe rs --namespace=sales-system -l app=sales

dev-events:
	kubectl get ev --sort-by metadata.creationTimestamp

dev-events-warn:
	kubectl get ev --field-selector type=Warning --sort-by metadata.creationTimestamp

dev-shell:
	kubectl exec --namespace=sales-system -it $(shell kubectl get pods --namespace=sales-system | grep sales | cut -c1-26) --container sales-api -- /bin/sh

dev-database-restart:
	kubectl rollout restart statefulset database --namespace=sales-system

# ==============================================================================
# Administration

migrate:
	go run app/tooling/sales-admin/main.go migrate

seed: migrate
	go run app/tooling/sales-admin/main.go seed

vault:
	go run app/tooling/sales-admin/main.go vault

token:
	go run app/tooling/sales-admin/main.go gentoken 5cf37266-3473-4006-984f-9325122678b7 54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

pgcli-local:
	pgcli postgresql://postgres:postgres@localhost

pgcli:
	pgcli postgresql://postgres:postgres@database-service.sales-system.svc.cluster.local

liveness-local:
	curl -il http://localhost:4000/debug/liveness

liveness:
	curl -il http://sales-service.sales-system.svc.cluster.local:4000/debug/liveness

readiness-local:
	curl -il http://localhost:4000/debug/readiness

readiness:
	curl -il http://sales-service.sales-system.svc.cluster.local:4000/debug/readiness

# ==============================================================================
# Metrics and Tracing

metrics-view-local-sc:
	expvarmon -ports="localhost:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view-sc:
	expvarmon -ports="sales-service.sales-system.svc.cluster.local:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view-local:
	expvarmon -ports="localhost:3001" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view:
	expvarmon -ports="sales-service.sales-system.svc.cluster.local:3001" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

zipkin-local:
	open -a "Google Chrome" http://localhost:9411/zipkin/

zipkin:
	open -a "Google Chrome" http://zipkin-service.sales-system.svc.cluster.local:9411/zipkin/

# ==============================================================================
# Running tests within the local computer
# go install honnef.co/go/tools/cmd/staticcheck@latest
# go install golang.org/x/vuln/cmd/govulncheck@latest

test:
	CGO_ENABLED=0 go test -count=1 ./...
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...
	govulncheck ./...

test-token-local:
	curl -il --user "admin@example.com:gophers" http://localhost:3000/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

test-token:
	curl -il --user "admin@example.com:gophers" http://sales-service.sales-system.svc.cluster.local:3000/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# export TOKEN="COPY TOKEN STRING FROM LAST CALL"

test-users-local:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users?page=1&rows=2

test-users:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://sales-service.sales-system.svc.cluster.local:3000/v1/users?page=1&rows=2

test-load-local:
	hey -m GET -c 100 -n 10000 -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users?page=1&rows=2

test-load:
	hey -m GET -c 100 -n 10000 -H "Authorization: Bearer ${TOKEN}" http://sales-service.sales-system.svc.cluster.local:3000/v1/users?page=1&rows=2

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-list:
	go list -m -u -mod=readonly all

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all
