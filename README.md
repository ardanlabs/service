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
* Vendoring dependencies with Modules, requires Go 1.11 or higher.

## Local Installation

This project contains three services and uses 3rd party services such as MongoDB and Zipkin. Docker is required to run this software on your local machine.

### Getting the project

You can use the traditional `go get` command to download this project into your configured GOPATH.

```
$ GO111MODULE=off go get -u gitHub.com/ardanlabs/service
```

### Go Modules

This project is using Go Module support for vendoring dependencies. We are using the `tidy` and `vendor` commands to maintain the dependencies and make sure the project can create reproducible builds. This project assumes the source code will be inside your GOPATH within the traditional location.

```
$ cd $GOPATH/src/github.com/ardanlabs/service
$ GO111MODULE=off go mod tidy
$ GO111MODULE=off go mod vendor
```

### Installing Docker

Docker is a critical component to managing and running this project. It kills me to just send you to the Docker installation page but it's all I got for now.

https://docs.docker.com/install/

If you are having problems installing docker reach out or jump on [Gopher Slack](http://invite.slack.golangbridge.org/) for help.

## Running The Project

All the source code, including any dependencies, have been vendored into the project. There is a single `dockerfile`and a `docker-compose` file that knows how to build and run all the services.

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

The service provides record keeping for someone running a multi-family garage sale. Authenticated users can maintain a list of products for sale.

<!--The service uses the following models:-->

<!--<img src="https://raw.githubusercontent.com/ardanlabs/service/master/models.jpg" alt="Garage Sale Service Models" title="Garage Sale Service Models" />-->

<!--(Diagram generated with draw.io using `models.xml` file)-->

### Making Requests

#### Initial User

To make a request to the service you must have an authenticated user. Users can be created with the API but an initial admin user must first be created. While the service is running you can create the initial user with the command `make admin`

```
$ make admin
```

This will create a user with email `admin@example.com` and password `gophers`.

#### Authenticating

Before any authenticated requests can be sent you must acquire an auth token. Make a request using HTTP Basic auth with your email and password to get the token.

```
$ curl --user "admin@example.com:gophers" http://localhost:3000/v1/users/token
```

I suggest putting the resulting token in an environment variable like `$TOKEN`.

#### Authenticated Requests

To make authenticated requests put the token in the `Authorization` header with the `Bearer ` prefix.

```
$ curl -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users
```

## What's Next

We are in the process of writing more documentation about this code. Classes are being finalized as part of the Ultimate series.
