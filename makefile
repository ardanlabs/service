SHELL := /bin/bash

all: keys sales-api metrics

keys:
	GO111MODULE=on go run -mod=vendor ./cmd/sales-admin/main.go keygen private.pem

admin:
	GO111MODULE=on go run -mod=vendor ./cmd/sales-admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

migrate:
	GO111MODULE=on go run -mod=vendor ./cmd/sales-admin/main.go --db-disable-tls=1 migrate

seed: migrate
	GO111MODULE=on go run -mod=vendor ./cmd/sales-admin/main.go --db-disable-tls=1 seed

sales-api:
	docker build \
		-t gcr.io/ardan-starter-kit/sales-api-amd64:1.0 \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

search:
	docker build \
		-t gcr.io/ardan-starter-kitsearch-amd64:1.0 \
		--build-arg PACKAGE_NAME=search \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

metrics:
	docker build \
		-t gcr.io/ardan-starter-kit/metrics-amd64:1.0 \
		--build-arg PACKAGE_NAME=metrics \
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
	GO111MODULE=on go test -mod=vendor ./...

clean:
	docker system prune -f

stop-all:
	docker stop $(docker ps -aq)

remove-all:
	docker rm $(docker ps -aq)

#===============================================================================
# GKE

config:
	@echo Setting environment for ardan-starter-kit
	gcloud config set project ardan-starter-kit
	gcloud config set compute/zone us-central1-b
	gcloud auth configure-docker
	@echo ======================================================================

project:
	gcloud projects create ardan-starter-kit
	gcloud beta billing projects link ardan-starter-kit --billing-account=$(ACCOUNT_ID)
	gcloud services enable container.googleapis.com
	@echo ======================================================================

cluster:
	gcloud container clusters create ardan-starter-cluster --num-nodes=2 --machine-type=n1-standard-2
	gcloud compute instances list
	@echo ======================================================================

upload:
	docker push gcr.io/ardan-starter-kit/sales-api-amd64:1.0
	docker push gcr.io/ardan-starter-kit/metrics-amd64:1.0
	@echo ======================================================================

database:
	gcloud sql instances create ardan-starter-db --database-version=POSTGRES_9_6 --no-backup --tier=db-f1-micro --zone=us-central1-b
	# https://console.cloud.google.com/sql/instances/ardan-starter-db/overview
	# Change Password
	# Whitelist IP address  IP/32
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

#===============================================================================
# GKE Installation
#
# Install the Google Cloud SDK. This contains the gcloud client needed to perform
# some operatings
# https://cloud.google.com/sdk/
#
# Installing the K8s kubectl client. 
# https://kubernetes.io/docs/tasks/tools/install-kubectl/