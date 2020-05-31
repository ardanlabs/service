SHELL := /bin/bash

export PROJECT = ardan-starter-kit

# Building containers

all: sales-api metrics

sales-api:
	docker build \
		-f dockerfile.sales-api \
		-t gcr.io/$(PROJECT)/sales-api-amd64:1.0 \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

metrics:
	docker build \
		-f dockerfile.metrics \
		-t gcr.io/$(PROJECT)/metrics-amd64:1.0 \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.

# Running from within docker compose

run: up seed

up:
	docker-compose up --detach --remove-orphans

down:
	docker-compose down --remove-orphans

logs:
	docker-compose logs -f

# Running from within the local computer

run-local: up-local seed

up-local:
	docker run -it -d -p 5432:5432 postgres:11.1-alpine

sales-local:
	cd cmd/sales-api; \
	go run main.go

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

# Administration

keys:
	go run cmd/sales-admin/main.go keygen private.pem

admin:
	go run cmd/sales-admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

migrate:
	go run cmd/sales-admin/main.go --db-disable-tls=1 migrate

seed: migrate
	go run cmd/sales-admin/main.go --db-disable-tls=1 seed

# Running tests within the local computer

test:
	go test ./... -count=1
	staticcheck ./...

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
	go mod vendor

deps-cleancache:
	go clean -modcache

# Docker support

clean:
	docker system prune -f
