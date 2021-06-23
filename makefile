SHELL := /bin/bash

export PROJECT = ardan-starter-kit

# ==============================================================================
# Testing running system

# For testing a simple query on the system. Don't forget to `make seed` first.
# curl --user "admin@example.com:gophers" http://localhost:3000/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1
# export TOKEN="COPY TOKEN STRING FROM LAST CALL"
# curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users/1/2

# For testing load on the service.
# hey -m GET -c 100 -n 10000 -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users/1/2
# zipkin: http://localhost:9411
# expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

# Used to install expvarmon program for metrics dashboard.
# go install github.com/divan/expvarmon@latest

# To generate a private/public key PEM file.
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# openssl rsa -pubout -in private.pem -out public.pem
# ./sales-admin genkey

# ==============================================================================
# Building containers

# $(shell git rev-parse --short HEAD)
VERSION := 1.0

all: sales metrics

sales:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api-amd64:$(VERSION) \
		--build-arg VCS_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

metrics:
	docker build \
		-f zarf/docker/dockerfile.metrics \
		-t metrics-amd64:$(VERSION) \
		--build-arg VCS_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/dev

KIND_CLUSTER := ardan-starter-cluster

# Upgrade to latest Kind (>=v0.11): e.g. brew upgrade kind
# For full Kind v0.11 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.11.0
# Kind release used for our project: https://github.com/kubernetes-sigs/kind/releases/tag/v0.11.1
# The image used below was copied by the above link and supports both amd64 and arm64.

kind-up:
	kind create cluster \
		--image kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6 \
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

kind-services:
	kustomize build zarf/k8s/kind/database-pod | kubectl apply -f -
	kubectl wait --namespace=database-system --timeout=120s --for=condition=Available deployment/database-pod
	kustomize build zarf/k8s/kind/sales-pod | kubectl apply -f -

kind-services-delete:
	kustomize build zarf/k8s/kind/sales-pod | kubectl delete -f -
	kustomize build zarf/k8s/kind/database-pod | kubectl delete -f -

kind-update: all kind-load
	kubectl rollout restart deployment sales-pod

kind-update-newversion: all kind-load kind-services

kind-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/logfmt/main.go

kind-logs-sales:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/logfmt/main.go -service=SALES-API | jq

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-describe:
	kubectl describe nodes
	kubectl describe svc
	kubectl describe pod -l app=sales

kind-describe-deployment:
	kubectl describe deployment sales-pod

kind-describe-replicaset:
	kubectl get rs
	kubectl describe rs -l app=sales

kind-events:
	kubectl get ev --sort-by metadata.creationTimestamp

kind-events-warn:
	kubectl get ev --field-selector type=Warning --sort-by metadata.creationTimestamp

kind-context-sales:
	kubectl config set-context --current --namespace=sales-system

kind-shell:
	kubectl exec -it $(shell kubectl get pods | grep app | cut -c1-26) --container app -- /bin/sh

kind-database:
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed

# ==============================================================================
# Administration

migrate:
	go run app/sales-admin/main.go migrate

seed: migrate
	go run app/sales-admin/main.go seed

# ==============================================================================
# Running tests within the local computer

test:
	go test ./... -count=1
	staticcheck -checks=all ./...

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
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all

# ==============================================================================
# Docker support

docker-down:
	docker rm -f $(shell docker ps -aq)

docker-clean:
	docker system prune -f	

docker-kind-logs:
	docker logs -f $(KIND_CLUSTER)-control-plane

# ==============================================================================
# GCP

export PROJECT = ardan-starter-kit
CLUSTER := ardan-starter-cluster
DATABASE := ardan-starter-db
ZONE := us-central1-b

gcp-config:
	@echo Setting environment for $(PROJECT)
	gcloud config set project $(PROJECT)
	gcloud config set compute/zone $(ZONE)
	gcloud auth configure-docker

gcp-project:
	gcloud projects create $(PROJECT)
	gcloud beta billing projects link $(PROJECT) --billing-account=$(ACCOUNT_ID)
	gcloud services enable container.googleapis.com

gcp-cluster:
	gcloud container clusters create $(CLUSTER) --enable-ip-alias --num-nodes=2 --machine-type=n1-standard-2
	gcloud compute instances list

gcp-upload:
	docker tag sales-api-amd64:1.0 gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION)
	docker tag metrics-amd64:1.0 gcr.io/$(PROJECT)/metrics-amd64:$(VERSION)
	docker push gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION)
	docker push gcr.io/$(PROJECT)/metrics-amd64:$(VERSION)

gcp-database:
	# Create User/Password
	gcloud beta sql instances create $(DATABASE) --database-version=POSTGRES_9_6 --no-backup --tier=db-f1-micro --zone=$(ZONE) --no-assign-ip --network=default
	gcloud sql instances describe $(DATABASE)

gcp-db-assign-ip:
	gcloud sql instances patch $(DATABASE) --authorized-networks=[$(PUBLIC-IP)/32]
	gcloud sql instances describe $(DATABASE)

gcp-db-private-ip:
	# IMPORTANT: Make sure you run this command and get the private IP of the DB.
	gcloud sql instances describe $(DATABASE)

gcp-services:
	kustomize build zarf/k8s/gcp/sales-pod | kubectl apply -f -

gcp-status:
	gcloud container clusters list
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch

gcp-status-full:
	kubectl describe nodes
	kubectl describe svc
	kubectl describe pod -l app=sales

gcp-events:
	kubectl get ev --sort-by metadata.creationTimestamp

gcp-events-warn:
	kubectl get ev --field-selector type=Warning --sort-by metadata.creationTimestamp

gcp-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/logfmt/main.go

gcp-logs-sales:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/logfmt/main.go -service=SALES-API | jq

gcp-shell:
	kubectl exec -it $(shell kubectl get pods | grep sales | cut -c1-26 | head -1) --container app -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed

gcp-delete-all: gcp-delete
	kustomize build zarf/k8s/gcp/sales-pod | kubectl delete -f -
	gcloud container clusters delete $(CLUSTER)
	gcloud projects delete sales-api
	gcloud container images delete gcr.io/$(PROJECT)/sales-api-amd64:$(VERSION) --force-delete-tags
	gcloud container images delete gcr.io/$(PROJECT)/metrics-amd64:$(VERSION) --force-delete-tags
	docker image remove gcr.io/sales-api/sales-api-amd64:$(VERSION)
	docker image remove gcr.io/sales-api/metrics-amd64:$(VERSION)

#===============================================================================
# GKE Installation
#
# Install the Google Cloud SDK. This contains the gcloud client needed to perform
# some operatings
# https://cloud.google.com/sdk/
#
# Installing the K8s kubectl client. 
# https://kubernetes.io/docs/tasks/tools/install-kubectl/