# Backend Microservices

## Made with
<p align="center">
  <img src="https://storage.googleapis.com/agungdp/static/logo/golang.png" width="80" alt="golang logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/docker.png" width="80" hspace="10" alt="docker logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/rest.png" width="80" hspace="10" alt="rest logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/graphql.png" width="80" alt="graphql logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/grpc.png" width="160" hspace="15" vspace="15" alt="grpc logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/kafka.png" height="80" alt="kafka logo" />
</p>

This repository explain implementation of Go for building multiple microservices using a single codebase. Using [Standard Golang Project Layout](https://github.com/golang-standards/project-layout) and [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

### Create new service
```
make init service={{service_name}} modules={{module_a}},{{module_b}}
```

### Run service
```
make run service={{service_name}}
```

### Add new modules in existing service
```
make add-module service={{service_name}} modules={{module_c}},{{module_d}}
```

### Create docker image a service
```
make docker service={{service_name}}
```

## Services

* [**Auth Service**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/auth-service)
* [**Line Chatbot**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/line-chatbot#line-chatbot-service)
* [**Notification Service**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/notification-service)
* [**Storage Service**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/storage-service)
* [**User Service**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/user-service)

## Todo
- [x] Add task queue worker like celery
