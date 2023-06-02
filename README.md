# Candi, a library and utilities for `Membangun Candi` project in Golang

<a href="https://codeclimate.com/github/golangid/candi/maintainability"><img src="https://api.codeclimate.com/v1/badges/38c8703e672eb53bea87/maintainability" /></a>
[![Build Status](https://github.com/golangid/candi/workflows/build/badge.svg)](https://github.com/golangid/candi/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/golangid/candi)](https://goreportcard.com/report/github.com/golangid/candi)
[![codecov](https://codecov.io/gh/golangid/candi/branch/master/graph/badge.svg)](https://codecov.io/gh/golangid/candi)
[![golang](https://img.shields.io/badge/golang%20%3E=-v1.18-green.svg?logo=go)](https://golang.org/doc/devel/release.html#go1.18)

## Build with :heart: and
<p align="center">
  <img src="https://storage.googleapis.com/agungdp/static/logo/golang.png" width="80" alt="golang logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/docker.png" width="80" hspace="10" alt="docker logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/rest.png" width="80" hspace="10" alt="rest logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/graphql.png" width="80" alt="graphql logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/grpc.png" width="160" hspace="15" vspace="15" alt="grpc logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/kafka.png" height="80" alt="kafka logo" />
</p>


## Install `CLI`
For linux:
```
$ wget -O candi https://storage.googleapis.com/agungdp/bin/candi/candi-linux && chmod +x candi && sudo mv candi /usr/local/bin
$ candi
```

For macOS:
```
$ wget -O candi https://storage.googleapis.com/agungdp/bin/candi/candi-osx && chmod +x candi && mv candi /usr/local/bin
$ candi
```

For windows:
```
https://storage.googleapis.com/agungdp/bin/candi/candi-x64.exe (64 bit)
or 
https://storage.googleapis.com/agungdp/bin/candi/candi-x86.exe (32 bit)
```

Or build binary from source:
```
$ go install github.com/golangid/candi/cmd/candi@latest
$ candi
```

Flag options:
```
$ candi --help
Usage of candi:
  -add-handler
        [project generator] add handler in delivery module in service
  -add-module
        [project generator] add module in service
  -init
        [project generator] init service
  -init-monorepo
        [project generator] init monorepo codebase
  -libraryname string
        [project generator] define library name (default "github.com/golangid/candi"), you can custom set to CANDI_CLI_PACKAGES global environment variable 
  -monorepo-name string
        [project generator] set monorepo project name (default "monorepo")
  -output string
        [project generator] directory to write project to (default is service name)
  -packageprefix string
        [project generator] define package prefix
  -protooutputpkg string
        [project generator] define generated proto output target (if using grpc), with prefix is your go.mod
  -run
        [service runner] run selected service or all service in monorepo
  -scope string
        [project generator] set scope 
        1 for init service, 
        2 for add module(s), 
        3 for add delivery handler(s) in module, 
        4 for init monorepo codebase, 
        5 for run multiple service in monorepo
  -service string
        Describe service name (if run multiple services, separate by comma)
  -version
        print version
  -withgomod
        [project generator] generate go.mod or not (default true)
```


## Create new service or add module in existing service
```
$ candi
```
![](https://storage.googleapis.com/agungdp/static/candi/candi.gif)

### The project is generated with this architecture diagram:
![](https://storage.googleapis.com/agungdp/static/candi/arch.jpg?11)


## Build and run service
```
$ cd {{service_name}}
$ make run
```
If include GRPC handler, run `$ make proto` for generate rpc files from proto (must install `protoc` compiler min version `libprotoc 3.14.0`)

## Server handlers example:
* [**Example REST API in delivery layer**](https://github.com/golangid/backend-microservices/tree/master/services/user-service/internal/modules/auth/delivery/resthandler)
* [**Example gRPC in delivery layer**](https://github.com/golangid/backend-microservices/blob/master/services/storage-service/internal/modules/storage/delivery/grpchandler/grpchandler.go)
* [**Example GraphQL in delivery layer**](https://github.com/golangid/backend-microservices/tree/master/services/user-service/internal/modules/auth/delivery/graphqlhandler)

## Worker handlers example:
* [**Example Cron worker in delivery layer**](https://github.com/golangid/candi/tree/master/codebase/app/cron_worker) (Static Scheduler)
* [**Example Kafka consumer in delivery layer**](https://github.com/golangid/candi/tree/master/codebase/app/kafka_worker) (Event Driven Handler)
* [**Example Redis subscriber in delivery layer**](https://github.com/golangid/candi/tree/master/codebase/app/redis_worker) (Dynamic Scheduler)
* [**Example Task queue worker in delivery layer**](https://github.com/golangid/candi/tree/master/codebase/app/task_queue_worker)
* [**Example Postgres event listener in delivery layer**](https://github.com/golangid/candi/tree/master/codebase/app/postgres_worker)
* [**Example RabbitMQ consumer in delivery layer**](https://github.com/golangid/candi/tree/master/codebase/app/rabbitmq_worker) (Event Driven Handler and Dynamic Scheduler)

## Plugin: [Candi Plugin](https://github.com/golangid/candi-plugin)

## Example microservices project using this library: [Backend Microservices](https://github.com/agungdwiprasetyo/backend-microservices)

## Features
- Tracing
  - Using Jaeger for trace distributed system in microservices.
![](https://storage.googleapis.com/agungdp/static/candi/jaeger_tracing.png)

- Graceful Shutdown for all servers and workers

## Todo
- [x] ~~Add task queue worker like celery and add UI for manage task queue worker~~ => https://github.com/agungdwiprasetyo/task-queue-worker-dashboard
- [ ] Add Documentation


## Pronunciation
`/can·di/` lihat! beliau membangun seribu candi dalam satu malam.

## Contributing

A big thank you to all contributors!

<a href="https://github.com/golangid/candi/graphs/contributors">
  <img src="https://contributors-img.web.app/image?repo=golangid/candi" />
</a>
