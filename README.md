# Backend Microservices

## Made with
<p align="center">
  <img src="https://storage.googleapis.com/agungdp/static/logo/golang.png" width="80" alt="golang logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/docker.png" width="80" hspace="10" alt="docker logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/graphql.png" width="80" alt="graphql logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/grpc.png" width="140" hspace="15" vspace="15" alt="grpc logo" />
  <img src="https://storage.googleapis.com/agungdp/static/logo/kafka.png" height="80" alt="kafka logo" />
</p>

This repository explain implementation of Go for building multiple microservices using a single codebase. Using [Standard Golang Project Layout](https://github.com/golang-standards/project-layout) and [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

### Create new service
```
go run cmd/scaffold_maker/*.go --servicename={{service_name}} --modules={{module_a}},{{module_b}}
```


### Run service
```
make run {{service_name}}
```


### Create docker image a service
```
make docker {{service_name}}
```

## Services

* [**Line Chatbot**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/line-chatbot#line-chatbot-service)
* [**Warung Service**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/warung)
* [**Wedding Service**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/wedding)
* [**CMS**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/cms)
* [**Storage Service**](https://github.com/agungdwiprasetyo/backend-microservices/tree/master/cmd/storage-service)
