SHELL := /bin/bash

all: crud metrics tracer

crud:
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	docker build \
		-t crud-amd64:1.0 \
		-f dockerfile.crud \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

metrics:
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	docker build \
		-t metrics-amd64:1.0 \
		-f dockerfile.metrics \
		--build-arg VCS_REF=`git rev-parse HEAD` \
		--build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		.
	docker system prune -f

tracer:
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	docker build \
		-t tracer-amd64:1.0 \
		-f dockerfile.tracer \
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