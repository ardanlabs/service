SHELL := /bin/bash

all: crud metrics tracer

crud:
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	docker build -t crud-amd64 -f dockerfile.crud .
	docker system prune -f

metrics:
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	docker build -t metrics-amd64 -f dockerfile.metrics .
	docker system prune -f

tracer:
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	docker build -t tracer-amd64 -f dockerfile.tracer .
	docker system prune -f

up:
	docker-compose up

down:
	docker-compose down

test:  
	cd "$$GOPATH/src/github.com/ardanlabs/service"
	go test ./...
