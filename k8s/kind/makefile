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
	kubectl create -f deploy-postgres.yaml
	kubectl create -f deploy-sales-api.yaml
	@echo ======================================================================

get-pods:
	kubectl get pods

status:
	kubectl get nodes
	kubectl get pods
	kubectl get services sales-api
	kubectl logs <POD_ID> -c sales-api
	@echo ======================================================================

shell:
	# kubectl get pods
	# kubectl exec -it <POD NAME> --container sales-api  -- /bin/sh
	# ./admin --db-disable-tls=1 migrate
	# ./admin --db-disable-tls=1 seed
	# curl --user "admin@example.com:gophers" http://localhost:3000/v1/users/token
	# export TOKEN="COPY TOKEN STRING FROM LAST CALL"
	# curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users
	@echo ======================================================================

delete:
	kubectl delete services sales-api
	kubectl delete deployment sales-api	
	@echo ======================================================================