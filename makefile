SHELL := /bin/bash

export PROJECT = ardan-starter-kit

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

run: up seed

up:
	docker-compose up --detach --remove-orphans

down:
	docker-compose down --remove-orphans

logs:
	docker-compose logs -f

keys:
	go run ./cmd/sales-admin/main.go keygen private.pem

admin:
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 useradd admin@example.com gophers

migrate:
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 migrate

seed: migrate
	go run ./cmd/sales-admin/main.go --db-disable-tls=1 seed

test:
	go test ./... -count=1
	staticcheck ./...

clean:
	docker system prune -f

stop-all:
	docker stop $(docker ps -aq)

remove-all:
	docker rm $(docker ps -aq)

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