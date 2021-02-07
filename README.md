# Candi, a framework for `Membangun Candi` project in Golang

<a href="https://codeclimate.com/github/agungdwiprasetyo/candi/maintainability"><img src="https://api.codeclimate.com/v1/badges/38c8703e672eb53bea87/maintainability" /></a>
[![Build Status](https://github.com/agungdwiprasetyo/candi/workflows/build/badge.svg)](https://github.com/agungdwiprasetyo/candi/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/agungdwiprasetyo/candi)](https://goreportcard.com/report/github.com/agungdwiprasetyo/candi)
[![codecov](https://codecov.io/gh/agungdwiprasetyo/candi/branch/master/graph/badge.svg)](https://codecov.io/gh/agungdwiprasetyo/candi)

## Made with
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
$ wget -O candi https://storage.googleapis.com/agungdp/bin/candi/candi-linux && chmod +x candi
$ ./candi
```

For macOS:
```
$ wget -O candi https://storage.googleapis.com/agungdp/bin/candi/candi-osx && chmod +x candi
$ ./candi
```

For windows:
```
https://storage.googleapis.com/agungdp/bin/candi/candi-x64.exe (64 bit)
or 
https://storage.googleapis.com/agungdp/bin/candi/candi-x86.exe (32 bit)
```

Or build binary from source:
```
$ go get -u pkg.agungdp.dev/candi/cmd/candi
$ candi
```

Flag options:
```
$ candi --help
Usage of candi:
  -libraryname string
        [project generator] define library name (default "pkg.agungdp.dev/candi")
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
        3 for init monorepo codebase, 
        4 for init service in monorepo, 
        5 for add module(s) service in monorepo, 
        6 for run multiple service in monorepo
  -service string
        [service runner] depend to "-run" flag, run specific services (if multiple services, separate by comma)
  -servicename string
        [project generator] define service name
  -withgomod
        [project generator] generate go.mod or not (default true)
```


## Create new service or add module in existing service
```
$ candi
```
![](https://storage.googleapis.com/agungdp/static/candi/candi.gif)

## Build and run service
```
$ cd {{service_name}}
$ make run
```
If include GRPC handler, run `$ make proto` for generate rpc files from proto (must install `protoc` compiler min version `libprotoc 3.14.0`)


## Service handlers example:
* [**Example Cron worker in delivery layer**](https://github.com/agungdwiprasetyo/candi/tree/master/codebase/app/cron_worker)
* [**Example Kafka consumer in delivery layer**](https://github.com/agungdwiprasetyo/candi/tree/master/codebase/app/kafka_worker)
* [**Example Redis subscriber in delivery layer**](https://github.com/agungdwiprasetyo/candi/tree/master/codebase/app/redis_worker)
* [**Example Task queue worker in delivery layer**](https://github.com/agungdwiprasetyo/candi/tree/master/codebase/app/task_queue_worker)


## Todo
- [x] ~~Add task queue worker like celery and add UI for manage task queue worker~~ => https://github.com/agungdwiprasetyo/task-queue-worker-dashboard
- [ ] Add Documentation
