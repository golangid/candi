# Candi, a framework for `Membangun Candi` project in Golang

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
```
$ go get -u pkg.agungdwiprasetyo.com/candi/cmd/candi
```
Flag options:
```
$ candi --help
Usage of candi:
  -libraryname string
        define library name (default "pkg.agungdwiprasetyo.com/candi")
  -output string
        directory to write project to (default is service name)
  -packageprefix string
        define package prefix
  -protooutputpkg string
        define generated proto output target (if using grpc), with prefix is your go.mod
  -scope string
        set scope (1 for init service, 2 for add module(s)
  -servicename string
        define service name
  -withgomod
        generate go.mod or not (default true)
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
