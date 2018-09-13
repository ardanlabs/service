# Ultimate Service

Copyright 2018, Ardan Labs  
info@ardanlabs.com

## Licensing

```
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## Description

Service is a project that provides a starter-kit for a REST based web service. It provides best practices around Go web services using POD architecture and design. It contains the following features:

* Minimal application web framework.
* Middleware integration.
* Database support using MongoDB.
* CRUD based pattern.
* Distributed logging and tracing.
* Testing patterns.
* User authentication.
* POD architecture with sidecars for metrics and tracing.
* Use of Docker, Docker Compose, and Makefiles.
* Vendoring dependencies with Modules, requires Go 1.11.
* Deployment to Azure using ACI.

## Local Installation

This project contains three services and uses 3rd party services such as MongoDB and Zipkin. Docker is required to run this software on your local machine.

### Go Modules

This project is using Go Module support starting with go1.11. By default modules are set to `auto`. If you have changed this, please make sure it is reset back to `auto` for this project.

```
export GO111MODULE=auto
```

We are using the `tidy` and `vendor` commands to maintain the dependencies and make sure the project can create reproducible builds.

```
go mod tidy
go mod vendor
```

### Getting the project

You can use the traditional `go get` command to download this project into your configured GOPATH.

```
$ go get -u github.com/ardanlabs/service
```

### Installing Docker

Docker is a critical component to managing and running this project. It kills me to just send you to the Docker installation page but it's all I got for now.

https://docs.docker.com/install/

If you are having problems installing docker reach out or jump on [Gopher Slack](http://invite.slack.golangbridge.org/) for help.

## Running The Project

All the source code, including any dependencies, have been vendored into the project. There is a single `dockerfile` for each service and a `docker-compose` file that knows how to run all the services.

A `makefile` has also been provide to make building, running and testing the software easier.

### Building the project

Navigate to the root of the project and use the `makefile` to build all of the services.

```
$ cd $GOPATH/src/github.com/ardanlabs/service
$ make all
```

### Running the project

Navigate to the root of the project and use the `makefile` to run all of the services.

```
$ cd $GOPATH/src/github.com/ardanlabs/service
$ make up
```

The `make up` command will leverage Docker Compose to run all the services, including the 3rd party services. The first time to run this command, Docker will download the required images for the 3rd party services.

Default configuration is set which should be valid for most systems. Use the `docker-compose.yaml` file to configure the services differently is necessary. Email me if you have issues or questions.

### Stopping the project

You can hit <ctrl>C in the terminal window running `make up`. Once that shutdown sequence is complete, it is important to run the `make down` command.

```
$ <ctrl>C
$ make down
```

Running `make down` will properly stop and terminate the Docker Compose session.

## About The Project

The service provides record keeping for someone running a multi-family garage sale. Authenticated users can maintain a list of products for sale, record transactions as items are sold, and generate payout amounts for each family.

The service uses the following models:

<img src="https://raw.githubusercontent.com/ardanlabs/service/master/models.jpg" alt="Garage Sale Service Models" title="Garage Sale Service Models" />

(Diagram generated with draw.io using `models.xml` file)

## What's Next

We are in the process of writing more documentation about this code. Classes are being finalized as part of the Ultimate series.
