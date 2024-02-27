# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# Deploy First Mentality

# ==============================================================================
# Go Installation
#
#   You need to have Go version 1.22 to run this code.
#
#   https://go.dev/dl/
#
#   If you are not allowed to update your Go frontend, you can install
#   and use a 1.22 frontend.
#
#   $ go install golang.org/dl/go1.22@latest
#   $ go1.22 download
#
#   This means you need to use `go1.22` instead of `go` for any command
#   using the Go frontend tooling from the makefile.

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
#   This project uses Docker and it is expected to be installed. Please provide
#   Docker at least 4 CPUs.
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
#

# ==============================================================================
# Running The Project
#
#	$ make dev-up
#	$ make dev-update-apply
#   $ make token
#   $ export TOKEN=<token>
#   $ make users
#
#   You can use `make dev-status` to look at the status of your KIND cluster.

# ==============================================================================
# Project Tooling
#
#   There is tooling that can generate documentation and add a new domain to
#   the code base. The code that is generated for a new domain provides the
#   common code needed for all domains.
#
#   Generating Documentation
#   $ go run app/tooling/docs/main.go --browser
#   $ go run app/tooling/docs/main.go -out json
#
#   Adding New Domain To System
#   $ go run app/tooling/sales-admin/main.go domain sale

# ==============================================================================
# CLASS NOTES
#
# Kind
# 	For full Kind v0.22 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.22.0
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
# Define dependencies

GOLANG          := golang:1.22
ALPINE          := alpine:3.19
KIND            := kindest/node:v1.29.2
POSTGRES        := postgres:16.2
GRAFANA         := grafana/grafana:10.3.0
PROMETHEUS      := prom/prometheus:v2.50.0
TEMPO           := grafana/tempo:2.3.0
LOKI            := grafana/loki:2.9.0
PROMTAIL        := grafana/promtail:2.9.0

KIND_CLUSTER    := ardan-starter-cluster
NAMESPACE       := sales-system
APP             := sales
BASE_IMAGE_NAME := ardanlabs/service
SERVICE_NAME    := sales-api
VERSION         := 0.0.1
SERVICE_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)
METRICS_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME)-metrics:$(VERSION)

# VERSION       := "0.0.1-$(shell git rev-parse --short HEAD)"

# ==============================================================================
# Install dependencies

dev-gotooling:
	go install github.com/divan/expvarmon@latest
	go install github.com/rakyll/hey@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest

dev-brew:
	brew update
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize
	brew list pgcli || brew install pgcli
	brew list watch || brew instal watch

dev-docker:
	docker pull $(GOLANG)
	docker pull $(ALPINE)
	docker pull $(KIND)
	docker pull $(POSTGRES)
	docker pull $(GRAFANA)
	docker pull $(PROMETHEUS)
	docker pull $(TEMPO)
	docker pull $(LOKI)
	docker pull $(PROMTAIL)

# ==============================================================================
# Building containers

all: service metrics

service:
	docker build \
		-f zarf/docker/dockerfile.service \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

metrics:
	docker build \
		-f zarf/docker/dockerfile.metrics \
		-t $(METRICS_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/kind

dev-up:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)
	kind load docker-image $(GRAFANA) --name $(KIND_CLUSTER)
	kind load docker-image $(PROMETHEUS) --name $(KIND_CLUSTER)
	kind load docker-image $(TEMPO) --name $(KIND_CLUSTER)
	kind load docker-image $(LOKI) --name $(KIND_CLUSTER)
	kind load docker-image $(PROMTAIL) --name $(KIND_CLUSTER)

dev-down:
	kind delete cluster --name $(KIND_CLUSTER)

dev-status-all:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-status:
	watch -n 2 kubectl get pods -o wide --all-namespaces

# ------------------------------------------------------------------------------

dev-load:
	kind load docker-image $(SERVICE_IMAGE) --name $(KIND_CLUSTER)
	kind load docker-image $(METRICS_IMAGE) --name $(KIND_CLUSTER)

# podman is currently experimental, and fails for some reason with kind load
# docker-image (possibly a tagging issue?) but the below works.
dev-load-podman:
	docker image save $(SERVICE_IMAGE):$(VERSION) -o image-sales
	kind load image-archive image-sales --name ardan-starter-cluster
	rm -f image-sales-metrics image-sales

	rm -f image-sales-metrics image-sales
	docker image save $(METRICS_IMAGE):$(VERSION) -o image-sales-metrics
	kind load image-archive image-sales-metrics --name ardan-starter-cluster

dev-apply:
	kustomize build zarf/k8s/dev/grafana | kubectl apply -f -
	kustomize build zarf/k8s/dev/prometheus | kubectl apply -f -
	kustomize build zarf/k8s/dev/tempo | kubectl apply -f -
	kustomize build zarf/k8s/dev/loki | kubectl apply -f -
	kustomize build zarf/k8s/dev/promtail | kubectl apply -f -

	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database

	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(APP) --timeout=120s --for=condition=Ready

dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)

dev-update: all dev-load dev-restart

dev-update-apply: all dev-load dev-apply

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 --max-log-requests=6 | go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

# ------------------------------------------------------------------------------

dev-logs-init:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) -f --tail=100 -c init-migrate-seed

dev-describe-node:
	kubectl describe node

dev-describe-deployment:
	kubectl describe deployment --namespace=$(NAMESPACE) $(APP)

dev-describe-sales:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(APP)

dev-describe-database:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=database

dev-describe-grafana:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=grafana

# ------------------------------------------------------------------------------

dev-logs-db:
	kubectl logs --namespace=$(NAMESPACE) -l app=database --all-containers=true -f --tail=100

dev-logs-grafana:
	kubectl logs --namespace=$(NAMESPACE) -l app=grafana --all-containers=true -f --tail=100

dev-logs-tempo:
	kubectl logs --namespace=$(NAMESPACE) -l app=tempo --all-containers=true -f --tail=100

dev-logs-loki:
	kubectl logs --namespace=$(NAMESPACE) -l app=loki --all-containers=true -f --tail=100

dev-logs-promtail:
	kubectl logs --namespace=$(NAMESPACE) -l app=promtail --all-containers=true -f --tail=100

# ------------------------------------------------------------------------------

dev-services-delete:
	kustomize build zarf/k8s/dev/sales | kubectl delete -f -
	kustomize build zarf/k8s/dev/grafana | kubectl delete -f -
	kustomize build zarf/k8s/dev/tempo | kubectl delete -f -
	kustomize build zarf/k8s/dev/loki | kubectl delete -f -
	kustomize build zarf/k8s/dev/promtail | kubectl delete -f -
	kustomize build zarf/k8s/dev/database | kubectl delete -f -

dev-describe-replicaset:
	kubectl get rs
	kubectl describe rs --namespace=$(NAMESPACE) -l app=$(APP)

dev-events:
	kubectl get ev --sort-by metadata.creationTimestamp

dev-events-warn:
	kubectl get ev --field-selector type=Warning --sort-by metadata.creationTimestamp

dev-shell:
	kubectl exec --namespace=$(NAMESPACE) -it $(shell kubectl get pods --namespace=$(NAMESPACE) | grep sales | cut -c1-26) --container sales-api -- /bin/sh

dev-database-restart:
	kubectl rollout restart statefulset database --namespace=$(NAMESPACE)

# ==============================================================================
# Administration

migrate:
	export SALES_DB_HOST_PORT=localhost; go run app/tooling/sales-admin/main.go migrate

seed: migrate
	export SALES_DB_HOST_PORT=localhost; go run app/tooling/sales-admin/main.go seed

pgcli:
	pgcli postgresql://postgres:postgres@localhost

liveness:
	curl -il http://localhost:3000/v1/liveness

readiness:
	curl -il http://localhost:3000/v1/readiness

token-gen:
	export SALES_DB_HOST_PORT=localhost; go run app/tooling/sales-admin/main.go gentoken 5cf37266-3473-4006-984f-9325122678b7 54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

docs:
	go run app/tooling/docs/main.go --browser

# ==============================================================================
# Metrics and Tracing

metrics-view-sc:
	expvarmon -ports="localhost:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

metrics-view:
	expvarmon -ports="localhost:3001" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

grafana:
	open -a "Google Chrome" http://localhost:3100/

statsviz:
	open -a "Google Chrome" http://localhost:4000/debug/statsviz

# ==============================================================================
# Running tests within the local computer

test-race:
	CGO_ENABLED=1 go test -race -count=1 ./...

test-only:
	CGO_ENABLED=0 go test -count=1 ./...

lint:
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...

vuln-check:
	govulncheck ./...

test: test-only lint vuln-check

test-race: test-race lint vuln-check

# ==============================================================================
# Hitting endpoints

token:
	curl -il \
	--user "admin@example.com:gophers" http://localhost:3000/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# export TOKEN="COPY TOKEN STRING FROM LAST CALL"

users:
	curl -il \
	-H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/v1/users?page=1&rows=2"

load:
	hey -m GET -c 100 -n 1000 \
	-H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/v1/users?page=1&rows=2"

otel-test:
	curl -il \
	-H "Traceparent: 00-918dd5ecf264712262b68cf2ef8b5239-896d90f23f69f006-01" \
	--user "admin@example.com:gophers" http://localhost:3000/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

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

# ==============================================================================
# Class Stuff

run:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

run-help:
	go run app/services/sales-api/main.go --help | go run app/tooling/logfmt/main.go

curl:
	curl -il http://localhost:3000/v1/hack

curl-auth:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/hackauth

load-hack:
	hey -m GET -c 100 -n 100000 "http://localhost:3000/v1/hack"

admin:
	go run app/tooling/sales-admin/main.go

ready:
	curl -il http://localhost:3000/v1/readiness

live:
	curl -il http://localhost:3000/v1/liveness

curl-create:
	curl -il -X POST \
	-H "Authorization: Bearer ${TOKEN}" \
	-H 'Content-Type: application/json' \
	-d '{"name":"bill","email":"b@gmail.com","roles":["ADMIN"],"department":"IT","password":"123","passwordConfirm":"123"}' \
	http://localhost:3000/v1/users

# ==============================================================================
# Talk commands

talk-up:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)	

talk-apply:
	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database

	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(APP) --timeout=120s --for=condition=Ready

talk-build: all dev-load talk-apply

talk-load:
	hey -m GET -c 10 -n 1000 -H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/v1/users?page=1&rows=2"

talk-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 --max-log-requests=6

talk-logs-cpu:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 --max-log-requests=6 | grep SCHED

talk-logs-mem:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 --max-log-requests=6 | grep "ms clock"

talk-describe:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(APP)

talk-metrics:
	expvarmon -ports="localhost:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

# ==============================================================================
# Admin Frontend

ADMIN_FRONTEND_PREFIX := ./app/frontends/admin

write-token-to-env:
	echo "VITE_SERVICE_API=http://localhost:3000/v1" > ${ADMIN_FRONTEND_PREFIX}/.env
	make token | grep -o '"ey.*"' | awk '{print "VITE_SERVICE_TOKEN="$$1}' >> ${ADMIN_FRONTEND_PREFIX}/.env

admin-gui-install:
	pnpm -C ${ADMIN_FRONTEND_PREFIX} install 

admin-gui-dev: admin-gui-install
	pnpm -C ${ADMIN_FRONTEND_PREFIX} run dev 

admin-gui-build: admin-gui-install
	pnpm -C ${ADMIN_FRONTEND_PREFIX} run build 

admin-gui-start-build: admin-gui-build
	pnpm -C ${ADMIN_FRONTEND_PREFIX} run preview 

admin-gui-run: write-token-to-env admin-gui-start-build
