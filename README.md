# Service

William Kennedy  
Ardan Labs  
bill@ardanlabs.com

Service is a project that provides a starter-kit for a REST based web service. It provides best practices around Go   web services using POD architecture and design. It contains the following features:

* Minimal application web framework
* Middleware integration
* Database support using MongoDB
* CRUD based pattern
* Distributed logging and tracing
* Testing patterns
* POD architecture with sidecars for metrics and tracing
* Use of Docker, Docker Compose, Makefile
* Vendoring with [dep](https://github.com/golang/dep) and [vgo](https://github.com/golang/vgo)
* Deployment to Kubernetes

## Local Installation

This project contains three services and uses 3rd party services such as MongoDB and Zipkin. Docker is required to run this software on your local machine.

### Getting the project

You can use the traditional `go get` command to download this project into your configured GOPATH.

```
$ go get -u github.com/ardanlabs/gotraining
```

### Installing Docker

Docker is a critical component to managing and running this project. It kills me to just send you to the Docker installation page but it's all I got for now.

https://docs.docker.com/install/

If you are having problems installing docker reach out or jump on [Gopher Slack](http://invite.slack.golangbridge.org/) for help.

## Running The Project

All the source code, including any dependencies, have been vendored into the project. We have been experimenting with `vgo` but `dep` is the offical vendoring tool being used. There is a single `dockerfile` for each service and a `docker-compose` file that knows how to run all the services.

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

## What's Next

We are in the process of writing more documentation about this code. Classes are being finalized as part of the Ultimate series.
