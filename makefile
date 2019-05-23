SHELL := /bin/bash

all: keys sales-api metrics tracer

keys:
	go run ./cmd/sales-admin/main.go --cmd keygen

admin:
	go run ./cmd/sales-admin/main.go --cmd useradd --user_email admin@example.com --user_password gophers

sales-api:
	docker build \
		-t gcr.io/sales-api/sales-api-amd64:1.0 \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

metrics:
	docker build \
		-t gcr.io/sales-api/metrics-amd64:1.0 \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

tracer:
	docker build \
		-t gcr.io/sales-api/tracer-amd64:1.0 \
		--build-arg PACKAGE_NAME=tracer \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

up:
	docker-compose up

down:
	docker-compose down

test:
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	go test ./...

clean:
	docker system prune -f

stop-all:
	docker stop $(docker ps -aq)

remove-all:
	docker rm $(docker ps -aq)

#===============================================================================
# GKE

config:
	@echo Setting environment for sales-api
	gcloud config set project sales-api
	gcloud config set compute/zone us-central1-b
	gcloud auth configure-docker
	@echo ======================================================================

project:
	gcloud projects create sales-api
	gcloud beta billing projects link sales-api --billing-account=$(ACCOUNT_ID)
	gcloud services enable container.googleapis.com
	@echo ======================================================================

cluster:
	gcloud container clusters create sales-api-cluster --num-nodes=2 --machine-type=n1-standard-2
	gcloud compute instances list
	@echo ======================================================================

upload:
	docker push gcr.io/sales-api/sales-api-amd64:1.0
	docker push gcr.io/sales-api/metrics-amd64:1.0
	docker push gcr.io/sales-api/tracer-amd64:1.0
	@echo ======================================================================

database:
	kubectl create -f gke-deploy-database.yaml
	kubectl expose -f gke-expose-database.yaml --type=LoadBalancer
	@echo ======================================================================

services:
	kubectl create -f gke-deploy-sales-api.yaml
	kubectl expose -f gke-expose-sales-api.yaml --type=LoadBalancer
	@echo ======================================================================

shell:
	kubectl exec -it pod-name --container name -- /bin/bash
	@echo ======================================================================

status:
	gcloud container clusters list
	kubectl get nodes
	kubectl get pods
	kubectl get services sales-api
	@echo ======================================================================

delete:
	kubectl delete services sales-api
	kubectl delete deployment sales-api	
	gcloud container clusters delete sales-api-cluster
	gcloud projects delete sales-api
	docker image remove gcr.io/sales-api/sales-api-amd64:1.0
	docker image remove gcr.io/sales-api/metrics-amd64:1.0
	docker image remove gcr.io/sales-api/tracer-amd64:1.0
	@echo ======================================================================