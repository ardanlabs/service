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
# curl -il --user "admin@example.com:gophers" http://sales-service.sales-system.svc.cluster.local:3000/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
# export TOKEN="COPY TOKEN STRING FROM LAST CALL"
# curl -il -H "Authorization: Bearer ${TOKEN}" http://sales-service.sales-system.svc.cluster.local:3000/v1/users/1/2
#
# For testing load on the service.
# hey -m GET -c 100 -n 10000 -H "Authorization: Bearer ${TOKEN}" http://sales-service.sales-system.svc.cluster.local:3000/v1/users/1/2
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
# export VAULT_ADDR='http://vault-service.sales-system.svc.cluster.local:8200'
# vault secrets list
# vault kv get secret/sales
# vault kv put secret/sales key="some data"
# curl -H "X-Vault-Token: myroot" -X GET http://vault-service.sales-system.svc.cluster.local:8200/v1/secret/data/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
# curl -H "X-Vault-Token: myroot" -H "Content-Type: application/json" -X POST -d '{"data":{"pk":"PEM"}}' http://vault-service.sales-system.svc.cluster.local:8200/v1/secret/data/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
#
# To show what calls are being made underneath to the proxy and checksum db.
# curl https://proxy.golang.org/github.com/ardanlabs/conf/@v/list
# curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/list
# curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.info
# curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.mod
# curl https://proxy.golang.org/github.com/ardanlabs/conf/v3/@v/v3.1.1.zip
# curl https://sum.golang.org/lookup/github.com/ardanlabs/conf/v3@v3.1.1
#
# OPA Playground
# https://play.openpolicyagent.org/

# ==============================================================================
# Install dependencies

dev.setup.mac.common:
	brew update
	brew tap hashicorp/tap
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize
	brew list pgcli || brew install pgcli
	brew list vault || brew install vault

dev.setup.mac: dev.setup.mac.common
	brew datawire/blackbird/telepresence || brew install datawire/blackbird/telepresence

dev.setup.mac.arm64: dev.setup.mac.common
	brew datawire/blackbird/telepresence-arm64 || brew install datawire/blackbird/telepresence-arm64

dev.docker:
	docker pull golang:1.19
	docker pull alpine:3.16
	docker pull kindest/node:v1.25.3
	docker pull postgres:15-alpine
	docker pull hashicorp/vault:1.12
	docker pull openzipkin/zipkin:2.23
	docker pull docker.io/datawire/tel2:2.8.5

# ==============================================================================
# Building containers

# $(shell git rev-parse --short HEAD)
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
POSTGRES := postgres:15-alpine
VAULT := hashicorp/vault:1.12
ZIPKIN := openzipkin/zipkin:2.23
TELEPRESENCE := docker.io/datawire/tel2:2.8.5

dev-up:
	kind create cluster \
		--image kindest/node:v1.25.3@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1 \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml
	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner
	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)
	kind load docker-image $(VAULT) --name $(KIND_CLUSTER)
	kind load docker-image $(ZIPKIN) --name $(KIND_CLUSTER)
	telepresence --context=kind-$(KIND_CLUSTER) helm install
	telepresence --context=kind-$(KIND_CLUSTER) connect

dev-down:
	telepresence quit -r -u
	kind delete cluster --name $(KIND_CLUSTER)

dev-load:
	cd zarf/k8s/dev/sales; kustomize edit set image sales-api-image=sales-api:$(VERSION)
	cd zarf/k8s/dev/sales; kustomize edit set image metrics-image=metrics:$(VERSION)
	kind load docker-image sales-api:$(VERSION) --name $(KIND_CLUSTER)
	kind load docker-image metrics:$(VERSION) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/database
	kustomize build zarf/k8s/dev/vault | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/vault
	kustomize build zarf/k8s/dev/zipkin | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/zipkin
	kustomize build zarf/k8s/dev/sales | kubectl apply -f -

dev-restart:
	kubectl rollout restart deployment sales --namespace=sales-system

dev-update: all dev-load dev-restart

dev-update-apply: all dev-load dev-apply

dev-logs:
	kubectl logs --namespace=sales-system -l app=sales --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go -service=SALES-API

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

# *** OTHER ****************************************************************

dev-logs-vault:
	kubectl logs --namespace=sales-system -l app=vault --all-containers=true -f --tail=100

dev-logs-db:
	kubectl logs --namespace=sales-system -l app=database --all-containers=true -f --tail=100

dev-logs-zipkin:
	kubectl logs --namespace=sales-system -l app=zipkin --all-containers=true -f --tail=100

# *** EXTRAS *******************************************************************

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

dev-database:
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
	pgcli postgresql://postgres:postgres@database-service.sales-system.svc.cluster.local

liveness:
	curl -il http://sales-service.sales-system.svc.cluster.local:4000/debug/liveness

readiness:
	curl -il http://sales-service.sales-system.svc.cluster.local:4000/debug/readiness

# ==============================================================================
# Metrics and Tracing

metrics-view:
	expvarmon -ports="sales-service.sales-system.svc.cluster.local:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view-sidecar:
	expvarmon -ports="sales-service.sales-system.svc.cluster.local:3001" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

zipkin:
	open -a "Google Chrome" http://zipkin-service.sales-system.svc.cluster.local:9411/zipkin/

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
