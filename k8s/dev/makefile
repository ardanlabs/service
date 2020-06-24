SHELL := /bin/bash

all:
	(cd ../.. && make all)

sales-api:
	(cd ../.. && make sales-api)

metrics:
	(cd ../.. && make metrics)

test:
	(cd ../.. && make test)

up:
	kind create cluster --name ardan-starter-cluster --config kind-config.yaml

down:
	kind delete cluster --name ardan-starter-cluster

load:
	kind load docker-image gcr.io/ardan-starter-kit/sales-api-amd64:1.0 --name ardan-starter-cluster
	kind load docker-image gcr.io/ardan-starter-kit/metrics-amd64:1.0 --name ardan-starter-cluster
	# kind load docker-image openzipkin/zipkin:2.11 --name ardan-starter-cluster
	# kind load docker-image postgres:11.1-alpine --name ardan-starter-cluster

services:
	kustomize build . | kubectl apply -f -
	@echo ======================================================================

update-sales-api:
	# Build a new version using 1.1
	kind load docker-image gcr.io/ardan-starter-kit/sales-api-amd64:1.1 --name ardan-starter-cluster
	kubectl set image pod <POD_NAME> sales-api=gcr.io/ardan-starter-kit/sales-api-amd64:1.1
	kubectl delete pod <POD_NAME>

get-pods:
	kubectl get pods

logs:
	kubectl logs -lapp=sales-api --all-containers=true

status:
	kubectl get nodes
	kubectl get pods
	kubectl get services sales-api
	@echo ======================================================================

shell:
	# kubectl get pods
	# kubectl exec -it <POD NAME> --container app -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed
	# curl --user "admin@example.com:gophers" http://localhost:3000/v1/users/token
	# export TOKEN="COPY TOKEN STRING FROM LAST CALL"
	# curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users
	@echo ======================================================================

delete:
	kustomize build . | kubectl delete -f -
	@echo ======================================================================
