SHELL := /bin/bash

# ==============================================================================
# Testing running system

# Deploy First Mentality
#
# Other commands to install.
# go install github.com/divan/expvarmon@latest
# go install github.com/rakyll/hey@latest
#
# For full Kind v0.17 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.17.0
#
# For testing a simple query on the system. Don't forget to `make seed` first.
# curl --user "admin@example.com:gophers" http://localhost:3000/v1/users/token
# export TOKEN="COPY TOKEN STRING FROM LAST CALL"
# curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users/1/2
#
# For testing load on the service.
# hey -m GET -c 100 -n 10000 -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users/1/2
#
# To generate a private/public key PEM file.
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# openssl rsa -pubout -in private.pem -out public.pem
# ./sales-admin genkey
#
# Testing coverage.
# go test -coverprofile p.out
# go tool cover -html p.out
#
# Vault Information.
# READ THIS: https://developer.hashicorp.com/vault/docs/concepts/tokens
# export VAULT_TOKEN=myroot
# export VAULT_ADDR='http://0.0.0.0:8200'
# vault secrets list
# vault kv get secret/sales
# vault kv put secret/sales key="some data"
#
# curl \
#   -H "X-Vault-Token: myroot" \
#   -X GET \
#   http://0.0.0.0:8200/v1/secret/data/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
# 
# curl \
#   -H "X-Vault-Token: myroot" \
#   -H "Content-Type: application/json" \
#   -X POST \
#   -d '{"data":{"pk":"PEM"}}' \
#   http://127.0.0.1:8200/v1/secret/data/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
#
# To show what calls are being made underneath to the proxy and checksum db.
# curl https://proxy.golang.org/github.com/ardanlabs/conf/@v/list
# curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/list
# curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.info
# curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.mod
# curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.zip
# curl https://sum.golang.org/lookup/github.com/ardanlabs/conf/v3@v3.1.1
#

# ==============================================================================
# Install dependencies

dev.setup.mac:
	brew update
	brew tap hashicorp/tap
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize
	brew list pgcli || brew install pgcli
	brew list vault || brew install vault

dev.docker:
	docker pull hashicorp/vault:1.12
	docker pull openzipkin/zipkin:2.23
	docker pull kindest/node:v1.25.2
	docker pull golang:1.19
	docker pull postgres:14-alpine
	docker pull alpine:3.16

# ==============================================================================
# Building containers

# $(shell git rev-parse --short HEAD)
VERSION := 1.0

all: sales metrics

sales:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

metrics:
	docker build \
		-f zarf/docker/dockerfile.metrics \
		-t metrics-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/kind

KIND_CLUSTER := ardan-starter-cluster
POSTGRES := postgres:14-alpine
VAULT := hashicorp/vault:1.12
ZIPKIN := openzipkin/zipkin:2.23

kind-up:
	kind create cluster \
		--image kindest/node:v1.25.2@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1 \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=sales-system

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load:
	cd zarf/k8s/kind/sales-pod; kustomize edit set image sales-api-image=sales-api-amd64:$(VERSION)
	cd zarf/k8s/kind/sales-pod; kustomize edit set image metrics-image=metrics-amd64:$(VERSION)
	kind load docker-image sales-api-amd64:$(VERSION) --name $(KIND_CLUSTER)
	kind load docker-image metrics-amd64:$(VERSION) --name $(KIND_CLUSTER)
	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)
	kind load docker-image $(VAULT) --name $(KIND_CLUSTER)
	kind load docker-image $(ZIPKIN) --name $(KIND_CLUSTER)

kind-apply:
	kustomize build zarf/k8s/kind/database-pod | kubectl apply -f -
	kubectl wait --timeout=120s --for=condition=Available deployment/database-pod
	kustomize build zarf/k8s/kind/vault-pod | kubectl apply -f -
	kubectl wait --timeout=120s --for=condition=Available deployment/vault-pod
	kustomize build zarf/k8s/kind/zipkin-pod | kubectl apply -f -
	kubectl wait --timeout=120s --for=condition=Available deployment/zipkin-pod
	kustomize build zarf/k8s/kind/sales-pod | kubectl apply -f -

kind-restart:
	kubectl rollout restart deployment sales-pod

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply

kind-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-describe:
	kubectl describe nodes
	kubectl describe svc

# *** SALES-SYSTEM *************************************************************

kind-logs-sales:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go -service=SALES-API

kind-logs-metrics:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go -service=METRICS

kind-status-sales:
	kubectl get pods -o wide --watch --namespace=sales-system

kind-describe-deployment:
	kubectl describe deployment sales-pod

kind-describe-sales:
	kubectl describe pod -l app=sales

kind-context-sales:
	kubectl config set-context --current --namespace=sales-system

# *** OTHER-POD ****************************************************************

kind-logs-vault:
	kubectl logs -l app=vault --all-containers=true -f --tail=100

kind-logs-db:
	kubectl logs -l app=database --all-containers=true -f --tail=100

kind-logs-zipkin:
	kubectl logs -l app=zipkin --all-containers=true -f --tail=100

# *** EXTRAS *******************************************************************

kind-services-delete:
	kustomize build zarf/k8s/kind/sales-pod | kubectl delete -f -
	kustomize build zarf/k8s/kind/zipkin-pod | kubectl delete -f -
	kustomize build zarf/k8s/kind/database-pod | kubectl delete -f -

kind-describe-replicaset:
	kubectl get rs
	kubectl describe rs -l app=sales

kind-events:
	kubectl get ev --sort-by metadata.creationTimestamp

kind-events-warn:
	kubectl get ev --field-selector type=Warning --sort-by metadata.creationTimestamp

kind-shell:
	kubectl exec -it $(shell kubectl get pods | grep sales | cut -c1-26) --container sales-api -- /bin/sh

kind-database:
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed

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

pgcli:
	pgcli postgresql://postgres:postgres@localhost

liveness:
	curl -il http://localhost:4000/debug/liveness

readiness:
	curl -il http://localhost:4000/debug/readiness

# ==============================================================================
# Metrics and Tracing

metrics-view:
	expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view-sidecar:
	expvarmon -ports=":3001" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

zipkin:
	open -a "Google Chrome" http://localhost:9411/zipkin/

# ==============================================================================
# Running tests within the local computer
# go install honnef.co/go/tools/cmd/staticcheck@latest
# go install golang.org/x/vuln/cmd/govulncheck@latest

test:
	go test -count=1 ./...
	staticcheck -checks=all ./...
	govulncheck ./...

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all
