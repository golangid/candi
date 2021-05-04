package main

const (
	dockerfileTemplate = `# Stage 1
FROM golang:1.16.2-alpine3.13 AS dependency_builder

WORKDIR /go/src
ENV GO111MODULE=on

RUN apk update
RUN apk add --no-cache bash ca-certificates git make

COPY go.mod .
COPY go.sum .

RUN go mod download

# Stage 2
FROM dependency_builder AS service_builder

WORKDIR /usr/app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -o bin

# Stage 3
FROM alpine:latest  

ARG BUILD_NUMBER
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
ENV BUILD_NUMBER=$BUILD_NUMBER

RUN mkdir -p /root/api
RUN mkdir -p /root/cmd/{{.ServiceName}}
RUN mkdir -p /root/config/key
COPY --from=service_builder /usr/app/bin bin
COPY --from=service_builder /usr/app/.env .env
COPY --from=service_builder /usr/app/api /root/api

ENTRYPOINT ["./bin"]
`

	makefileTemplate = `.PHONY : build run

proto:
	$(foreach proto_file, $(shell find api/proto -name '*.proto'),\
	protoc --proto_path=api/proto --go_out=plugins=grpc:api/proto \
	--go_opt=paths=source_relative $(proto_file);)

migration:
	@go run cmd/migration/migration.go -dbconn "$(dbconn)"

build:
	go build -o bin

run: build
	./bin

docker:
	docker build -t {{.ServiceName}}:latest .

run-container:
	docker run --name={{.ServiceName}} --network="host" -d {{.ServiceName}}

# unit test & calculate code coverage
test:
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo ">> running unit test and calculate coverage"
	@go test ./... -cover -coverprofile=coverage.txt -covermode=count -coverpkg=$(PACKAGES)
	@go tool cover -func=coverage.txt

clear:
	rm bin {{.ServiceName}}
`

	gomodTemplate = `module {{.ServiceName}}

go 1.16

require pkg.agungdp.dev/candi {{.Version}}
`

	gitignoreTemplate = `bin
vendor
main_service.go
{{.ServiceName}}
coverage.txt
`

	jsonSchemaTemplate = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"$id": "example",
	"title": "json schema type",
	"type": "object",
	"properties": {}
}
`

	readmeTemplate = "# {{upper .ServiceName}}\n\n" +
		"## Build and run service\n" +
		"If include GRPC handler, run this command (must install `protoc` compiler min version `libprotoc 3.14.0`):\n\n" +
		"```\n" +
		"$ make proto\n" +
		"```\n\n" +
		"If using SQL database, run this command for migration:\n" +
		"```\n" +
		"$ make migration dbconn=\"(YOUR DATABASE URL CONNECTION)\"\n" +
		"```\n\n" +
		"And then, build and run this service:\n" +
		"```\n" +
		"$ make run\n" +
		"```\n\n" +
		"## Run unit test & calculate code coverage\n" +
		"```\n" +
		"$ make test\n" +
		"```\n\n" +
		"## Create docker image\n" +
		"```\n" +
		"$ make docker\n" +
		"```\n"

	readmeMonorepoTemplate = "# Backend Microservices\n\n" +
		"## Made with\n" +
		`<p align="center">` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/golang.png" width="80" alt="golang logo" />\` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/docker.png" width="80" hspace="10" alt="docker logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/rest.png" width="80" hspace="10" alt="rest logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/graphql.png" width="80" alt="graphql logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/grpc.png" width="160" hspace="15" vspace="15" alt="grpc logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/kafka.png" height="80" alt="kafka logo" />` + "\n" +
		"</p>\n\n" +
		"This repository explain implementation of Go for building multiple microservices using a single codebase. Using [Standard Golang Project Layout](https://github.com/golang-standards/project-layout) and [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)\n\n" +
		"## Create new service (for new project)\n" +
		"Please install **latest** [**candi**](https://pkg.agungdp.dev/candi) CLI first, and then:\n" +
		"```\n" +
		"$ candi -init\n" +
		"```\n" +
		"If include GRPC handler, run this command (must install `protoc` compiler min version `libprotoc 3.14.0`):\n\n" +
		"```\n" +
		"$ make proto service={{service_name}}\n" +
		"```\n\n" +
		"If using SQL database, run this command for migration:\n" +
		"```\n" +
		"$ make migration service={{service_name}} dbconn=\"{{YOUR DATABASE URL CONNECTION}}\"\n" +
		"```\n\n" +
		"## Run all services\n" +
		"```\n" +
		"$ candi -run\n" +
		"```\n\n" +
		"## Run specific service or multiple services\n" +
		"```\n" +
		"$ candi -run -service {{service_a}},{{service_b}}\n" +
		"```\n\n" +
		"## Add module(s) in specific service (project)\n" +
		"```\n" +
		"$ candi -add-module -service {{service_name}}\n" +
		"```\n\n" +

		"## Run unit test and calculate code coverage\n" +
		"* **Generate mocks first (using [mockery](https://github.com/vektra/mockery)):**\n" +
		"```\n" +
		"$ make mocks service={{service_name}}\n" +
		"```\n" +
		"* **Run test:**\n" +
		"```\n" +
		"$ make test service={{service_name}}\n" +
		"```\n" +

		"## Run sonar scanner\n" +
		"```\n" +
		"$ make sonar level={{level}} service={{service_name}}\n" +
		"```\n" +
		"`{{level}}` is service environment, example: `dev`, `staging`, or `prod`\n\n" +

		"## Create docker image a service\n" +
		"```\n" +
		"$ make docker service={{service_name}}\n" +
		"```\n"

	dockerfileMonorepoTemplate = `# Stage 1
FROM golang:1.16.2-alpine3.13 AS dependency_builder

WORKDIR /go/src
ENV GO111MODULE=on

RUN apk update
RUN apk add --no-cache bash ca-certificates git

COPY go.mod .
COPY go.sum .

RUN go mod download

# Stage 2
FROM dependency_builder AS service_builder

ARG SERVICE_NAME
WORKDIR /usr/app

COPY sdk sdk
COPY services/$SERVICE_NAME services/$SERVICE_NAME
COPY go.mod .
COPY go.sum .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -o bin services/$SERVICE_NAME/*.go

# Stage 3
FROM alpine:latest  

ARG BUILD_NUMBER
ARG SERVICE_NAME
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
ENV WORKDIR=services/$SERVICE_NAME/
ENV BUILD_NUMBER=$BUILD_NUMBER

RUN mkdir -p /root/services/$SERVICE_NAME
RUN mkdir -p /root/services/$SERVICE_NAME/api
RUN mkdir -p /root/services/$SERVICE_NAME/api/configs
COPY --from=service_builder /usr/app/bin bin
COPY --from=service_builder /usr/app/services/$SERVICE_NAME/.env /root/services/$SERVICE_NAME/.env
COPY --from=service_builder /usr/app/services/$SERVICE_NAME/api /root/services/$SERVICE_NAME/api
COPY --from=service_builder /usr/app/services/$SERVICE_NAME/configs /root/services/$SERVICE_NAME/configs

ENTRYPOINT ["./bin"]
`

	makefileMonorepoTemplate = `.PHONY : prepare build run

$(eval $(service):;@:)

check:
	@[ "${service}" ] || ( echo "\x1b[31;1mERROR: 'service' is not set\x1b[0m"; exit 1 )
	@if [ ! -d "services/$(service)" ]; then  echo "\x1b[31;1mERROR: service '$(service)' undefined\x1b[0m"; exit 1; fi

prepare: check
	@if [ ! -f services/$(service)/.env ]; then cp services/$(service)/.env.sample services/$(service)/.env; fi;

init:
	@candi -scope=4

add-module: check
	@candi -scope=5 -servicename=$(service)

proto: check
	@if [ ! -d "sdk/$(service)" ]; then echo "creating new proto files..." &&  mkdir sdk/$(service) && mkdir sdk/$(service)/proto; fi
	$(foreach proto_file, $(shell find services/$(service)/api/proto -name '*.proto'),\
	protoc --proto_path=services/$(service)/api/proto --go_out=plugins=grpc:sdk/$(service)/proto \
	--go_opt=paths=source_relative $(proto_file);)

migration: check
	@go run services/$(service)/cmd/migration/migration.go -dbconn "$(dbconn)"

build: check
	@go build -o services/$(service)/bin services/$(service)/*.go

run: build
	@WORKDIR="services/$(service)/" ./services/$(service)/bin

docker: check
	docker build --build-arg SERVICE_NAME=$(service) -t $(service):latest .

run-container:
	docker run --name=$(service) --network="host" -d $(service)

# mocks all interfaces in sdk for unit test
mocks:
	@mockery --all --keeptree --dir=sdk --output=./sdk/mocks
	@if [ -f sdk/mocks/Option.go ]; then rm sdk/mocks/Option.go; fi;

# unit test & calculate code coverage from selected service (please run mocks before run this rule)
test: check
	@echo "\x1b[32;1m>>> running unit test and calculate coverage for service $(service)\x1b[0m"
	@if [ -f services/$(service)/coverage.txt ]; then rm services/$(service)/coverage.txt; fi;
	@go test -race ./services/$(service)/... -cover -coverprofile=services/$(service)/coverage.txt -covermode=atomic \
		-coverpkg=$$(go list ./services/$(service)/... | grep -v -e mocks -e codebase | tr '\n' ',')
	@go tool cover -func=services/$(service)/coverage.txt

sonar: check
	@echo "\x1b[32;1m>>> running sonar scanner for service $(service)\x1b[0m"
	@[ "${level}" ] || ( echo "\x1b[31;1mERROR: 'level' is not set\x1b[0m" ; exit 1 )
	@sonar-scanner -Dsonar.projectKey=$(service)-$(level) \
	-Dsonar.projectName=$(service)-$(level) \
	-Dsonar.sources=./services/$(service) \
	-Dsonar.exclusions=**/mocks/**,**/module.go \
	-Dsonar.test.inclusions=**/*_test.go \
	-Dsonar.test.exclusions=**/mocks/** \
	-Dsonar.coverage.exclusions=**/mocks/**,**/*_test.go,**/main.go \
	-Dsonar.go.coverage.reportPaths=./services/$(service)/coverage.txt

clear:
	rm -rf sdk/mocks services/$(service)/mocks services/$(service)/bin bin
`

	gitignoreMonorepoTemplate = `vendor
bin
.env
*.pem
*.key
.vscode/
__debug_bin
coverage.txt
mocks/
.DS_Store
.scannerwork/
`
)
