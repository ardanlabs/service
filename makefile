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

# // To generate a private/public key PEM file.
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# openssl rsa -pubout -in private.pem -out public.pem
# ./sales-admin genkey

# ==============================================================================
# Building containers

all: sales metrics

sales:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api-amd64:1.0 \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f zarf/docker/dockerfile.metrics \
		-t metrics-amd64:1.0 \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

# ==============================================================================
# Running from within k8s/dev

dev-up:
	kind create cluster --image kindest/node:v1.21.1 --name ardan-starter-cluster --config zarf/k8s/dev/kind-config.yaml

dev-up-m1:
	kind create cluster --image rossgeorgiev/kind-node-arm64 --name ardan-starter-cluster --config zarf/k8s/dev/kind-config.yaml

dev-down:
	kind delete cluster --name ardan-starter-cluster

dev-load:
	kind load docker-image sales-api-amd64:1.0 --name ardan-starter-cluster
	kind load docker-image metrics-amd64:1.0 --name ardan-starter-cluster

dev-services:
	kustomize build zarf/k8s/dev | kubectl apply -f -

dev-update: sales
	kind load docker-image sales-api-amd64:1.0 --name ardan-starter-cluster
	kubectl delete pods -l app=sales

dev-metrics: metrics
	kind load docker-image metrics-amd64:1.0 --name ardan-starter-cluster
	kubectl delete pods -l app=sales

dev-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/logfmt/main.go

dev-logs-sales:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/logfmt/main.go -service=SALES-API | jq

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch

dev-status-full:
	kubectl describe nodes
	kubectl describe svc
	kubectl describe pod -l app=sales

dev-events:
	kubectl get ev --sort-by metadata.creationTimestamp

dev-events-warn:
	kubectl get ev --field-selector type=Warning --sort-by metadata.creationTimestamp

dev-shell:
	kubectl exec -it $(shell kubectl get pods | grep app | cut -c1-26) --container app -- /bin/sh

dev-database:
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed

dev-delete:
	kustomize build zarf/k8s/dev | kubectl delete -f -

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

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

clean:
	docker system prune -f	

logs-local:
	docker logs -f $(FILES)

# ==============================================================================
# GCP

export PROJECT = ardan-starter-kit
CLUSTER = ardan-starter-cluster
DATABASE = ardan-starter-db
ZONE = us-central1-b

stg-config:
	@echo Setting environment for $(PROJECT)
	gcloud config set project $(PROJECT)
	gcloud config set compute/zone $(ZONE)
	gcloud auth configure-docker

stg-project:
	gcloud projects create $(PROJECT)
	gcloud beta billing projects link $(PROJECT) --billing-account=$(ACCOUNT_ID)
	gcloud services enable container.googleapis.com

stg-cluster:
	gcloud container clusters create $(CLUSTER) --enable-ip-alias --num-nodes=2 --machine-type=n1-standard-2
	gcloud compute instances list

stg-upload:
	docker tag sales-api-amd64:1.0 gcr.io/$(PROJECT)/sales-api-amd64:1.0
	docker tag metrics-amd64:1.0 gcr.io/$(PROJECT)/metrics-amd64:1.0
	docker push gcr.io/$(PROJECT)/sales-api-amd64:1.0
	docker push gcr.io/$(PROJECT)/metrics-amd64:1.0

stg-database:
	# Create User/Password
	gcloud beta sql instances create $(DATABASE) --database-version=POSTGRES_9_6 --no-backup --tier=db-f1-micro --zone=$(ZONE) --no-assign-ip --network=default
	gcloud sql instances describe $(DATABASE)

stg-db-assign-ip:
	gcloud sql instances patch $(DATABASE) --authorized-networks=[$(PUBLIC-IP)/32]
	gcloud sql instances describe $(DATABASE)

stg-db-private-ip:
	# IMPORTANT: Make sure you run this command and get the private IP of the DB.
	gcloud sql instances describe $(DATABASE)

stg-services:
	kustomize build zarf/k8s/stg | kubectl apply -f -

stg-status:
	gcloud container clusters list
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch

stg-status-full:
	kubectl describe nodes
	kubectl describe svc
	kubectl describe pod -l app=sales

stg-events:
	kubectl get ev --sort-by metadata.creationTimestamp

stg-events-warn:
	kubectl get ev --field-selector type=Warning --sort-by metadata.creationTimestamp

stg-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/logfmt/main.go

stg-logs-sales:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/logfmt/main.go -service=SALES-API | jq

stg-shell:
	kubectl exec -it $(shell kubectl get pods | grep sales | cut -c1-26 | head -1) --container app -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed

stg-delete-all: stg-delete
	kustomize build zarf/k8s/dev | kubectl delete -f -
	gcloud container clusters delete $(CLUSTER)
	gcloud projects delete sales-api
	gcloud container images delete gcr.io/$(PROJECT)/sales-api-amd64:1.0 --force-delete-tags
	gcloud container images delete gcr.io/$(PROJECT)/metrics-amd64:1.0 --force-delete-tags
	docker image remove gcr.io/sales-api/sales-api-amd64:1.0
	docker image remove gcr.io/sales-api/metrics-amd64:1.0

#===============================================================================
# GKE Installation
#
# Install the Google Cloud SDK. This contains the gcloud client needed to perform
# some operatings
# https://cloud.google.com/sdk/
#
# Installing the K8s kubectl client. 
# https://kubernetes.io/docs/tasks/tools/install-kubectl/