SHELL := /bin/bash

export PROJECT = ardan-starter-kit

# ==============================================================================
# Building containers

all: sales-api metrics

sales-api:
	docker build \
		-f z/compose/dockerfile.sales-api \
		-t gcr.io/$(PROJECT)/sales-api-amd64:1.0 \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f z/compose/dockerfile.metrics \
		-t gcr.io/$(PROJECT)/metrics-amd64:1.0 \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

# ==============================================================================
# Running from within docker compose

run: up seed

up:
	docker-compose -f z/compose/docker-compose.yaml up --detach --remove-orphans

down:
	docker-compose -f z/compose/docker-compose.yaml down --remove-orphans

logs:
	docker-compose -f z/compose/docker-compose.yaml logs -f

# ==============================================================================
# Running from within k8s/dev

kind-up:
	kind create cluster --name ardan-starter-cluster --config z/k8s/dev/kind-config.yaml

kind-down:
	kind delete cluster --name ardan-starter-cluster

kind-load:
	kind load docker-image gcr.io/ardan-starter-kit/sales-api-amd64:1.0 --name ardan-starter-cluster
	kind load docker-image gcr.io/ardan-starter-kit/metrics-amd64:1.0 --name ardan-starter-cluster
	# kind load docker-image openzipkin/zipkin:2.11 --name ardan-starter-cluster
	# kind load docker-image postgres:11.1-alpine --name ardan-starter-cluster

kind-services:
	kustomize build z/k8s/dev | kubectl apply -f -
	@echo ======================================================================

kind-update-sales-api:
	# Build a new version using 1.1
	kind load docker-image gcr.io/ardan-starter-kit/sales-api-amd64:1.1 --name ardan-starter-cluster
	kubectl set image pod <POD_NAME> sales-api=gcr.io/ardan-starter-kit/sales-api-amd64:1.1
	kubectl delete pod <POD_NAME>

kind-logs:
	kubectl logs -lapp=sales-api --all-containers=true -f

kind-status:
	kubectl get nodes
	kubectl get pods
	kubectl get services sales-api
	@echo ======================================================================

kind-shell:
	# kubectl get pods
	# kubectl exec -it <POD NAME> --container app -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed
	# curl --user "admin@example.com:gophers" http://localhost:3000/v1/users/token
	# export TOKEN="COPY TOKEN STRING FROM LAST CALL"
	# curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users
	@echo ======================================================================

kind-delete:
	kustomize build . | kubectl delete -f -
	@echo ======================================================================

# ==============================================================================
# Administration

migrate:
	go run app/sales-admin/main.go --db-disable-tls=1 migrate

seed: migrate
	go run app/sales-admin/main.go --db-disable-tls=1 seed

# ==============================================================================
# Running tests within the local computer

test:
	go test ./... -count=1
	staticcheck ./...

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
	docker push gcr.io/$(PROJECT)/sales-api-amd64:1.0
	docker push gcr.io/$(PROJECT)/metrics-amd64:1.0

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
	# These scripts needs to be edited for the PROJECT and PRIVATE_DB_IP markers before running.
	kubectl create -f deploy-sales-api.yaml
	kubectl expose -f expose-sales-api.yaml --type=LoadBalancer

gcp-status:
	gcloud container clusters list
	kubectl get nodes
	kubectl get pods
	kubectl get services sales-api

gcp-shell:
	# kubectl get pods
	kubectl exec -it <POD NAME> --container sales-api  -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed

gcp-delete:
	kubectl delete services sales-api
	kubectl delete deployment sales-api	
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