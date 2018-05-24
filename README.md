# Service

Service is a project that provides a starter-kit for a REST based web service. It provides best practices around Go based web services using POD architecture and design. It contains the following features:

* Minimal application web framework
* Middleware integration
* DB database support using MongoDB
* CRUD based pattern
* Distributed logging and tracing
* Testing patterns
* POD architecture with sidecars for metrics and tracing
* Docker, Docker Compose, Makefile
* Vendoring with [dep](https://github.com/golang/dep) and [vgo](https://github.com/golang/vgo)

## Installation

There are two dockerfiles for both services at the root of the repo. Instructions for building the images are located in each respective dockerfile. Use docker-compose to run the services. The default configuration setting are for running with docker-compose.