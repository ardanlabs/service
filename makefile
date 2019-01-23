SHELL := /bin/bash

all: keys sales-api metrics tracer

keys:
	go run ./cmd/sales-admin/main.go --cmd keygen

admin:
	go run ./cmd/sales-admin/main.go --cmd useradd --user_email admin@example.com --user_password gophers

sales-api:
	docker build \
		-t sales-api-amd64:1.0 \
		--build-arg PACKAGE_NAME=sales-api \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

metrics:
	docker build \
		-t metrics-amd64:1.0 \
		--build-arg PACKAGE_NAME=metrics \
		--build-arg PACKAGE_PREFIX=sidecar/ \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

tracer:
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	docker build \
		-t tracer-amd64:1.0 \
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
