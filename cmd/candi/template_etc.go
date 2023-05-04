package main

const (
	dockerfileTemplate = `# Stage 1
FROM golang:1.18.4-alpine3.16 AS dependency_builder

WORKDIR /go/src
ENV GO111MODULE=on

RUN apk update
RUN apk add --no-cache bash ca-certificates git gcc musl-dev

COPY go.mod .
COPY go.sum .

RUN go mod download

# Stage 2
FROM dependency_builder AS service_builder

WORKDIR /usr/app

COPY . .
RUN go build -o bin

# Stage 3
FROM alpine:latest  

ARG BUILD_NUMBER
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /usr/app/
ENV BUILD_NUMBER=$BUILD_NUMBER

RUN mkdir -p /usr/app/api
COPY --from=service_builder /usr/app/bin bin
COPY --from=service_builder /usr/app/.env .env
COPY --from=service_builder /usr/app/api /usr/app/api

ENTRYPOINT ["./bin"]
`

	makefileTemplate = `.PHONY : build run

args = ` + "`arg=\"$(filter-out $@,$(MAKECMDGOALS))\" && echo $${arg:-${1}}`" + `

proto:
	$(foreach proto_file, $(shell find api/proto -name '*.proto'),\
	protoc --proto_path=api/proto --go_out=plugins=grpc:api/proto \
	--go_opt=paths=source_relative $(proto_file);)

migration:
	@go run cmd/migration/migration.go $(call args,up)

build:
	@go build -o bin

run: build
	@./bin

docker:
	docker build -t {{.ServiceName}}:latest .

run-container:
	docker run --name={{.ServiceName}} --network="host" -d {{.ServiceName}}

clear-docker:
	docker rm -f {{.ServiceName}}
	docker rmi -f {{.ServiceName}}

# mocks all interfaces for unit test
mocks:
	@mockery --all --keeptree --dir=internal --output=pkg/mocks --case underscore
	@mockery --all --keeptree --dir=pkg --output=pkg/mocks --case underscore

# unit test & calculate code coverage
test:
	@echo "\x1b[32;1m>>> running unit test and calculate coverage\x1b[0m"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@go test ./... -cover -coverprofile=coverage.txt -covermode=count \
		-coverpkg=$$(go list ./... | grep -v mocks | tr '\n' ',')
	@go tool cover -func=coverage.txt
`

	gomodTemplate = `module {{.ServiceName}}

go {{.GoVersion}}

require (
	{{.LibraryName}} {{.Version}}
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
)
`

	gitignoreTemplate = `bin
vendor
main_service.go
{{.ServiceName}}
coverage.txt
`

	jsonSchemaFilterGetTemplate = `{
	"$schema": "http://json-schema.org/draft-07/schema",
	"title": "JSON Schema for get all",
	"type": "object",
	"properties": {
		"page": {
			"type": "number",
			"default": 1,
			"minimum": 0
		},
		"limit": {
			"type": "number",
			"default": 10,
			"minimum": 1
		},
		"orderBy": {
			"type": "string",
			"enum": ["id", "field", "createdAt", "updatedAt"]
		},
		"sort": {
			"type": "string",
			"enum": ["asc", "desc", "ASC", "DESC"]
		},
		"search": {
			"type": "string"
		},
		"startDate": {
			"anyOf": [
				{
					"type": "string",
					"format": "date"
				},
				{
					"type": "string",
					"maxLength": 0
				}
			]
		},
		"endDate": {
			"anyOf": [
				{
					"type": "string",
					"format": "date"
				},
				{
					"type": "string",
					"maxLength": 0
				}
			]
		}
	},
	"additionalProperties": true
}
`

	jsonSchemaSaveTemplate = `{
	"$schema": "http://json-schema.org/draft-07/schema",
	"title": "JSON Schema for save",
	"type": "object",
	"properties": {
		"id": {
			"type": "{{if and .MongoDeps (not .SQLDeps)}}string{{else}}integer{{end}}"
		},
		"field": {
			"type": "string",
			"minLength": 1
		}
	},
	"required": [ "field" ],
	"additionalProperties": false
}
`

	mitLicenseTemplate = `The MIT License (MIT)

Copyright (c) {{.Year}} {{.Owner}}

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.`

	apacheLicenseTemplate = `Copyright {{.Year}} {{.Owner}}

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.`

	privateLicenseTemplate = `Copyright {{.Year}} {{.Owner}}

NOTICE:

All Information contained herein is, and remains the property of
{{.Owner}} and its suppliers, if any. The intellectual
and technical concepts contained herein are proprietary to 
{{.Owner}} and its suppliers and maybe covered by Republic of
Indonesia and Foreign Patents, patents in process, and are protected by
trade secret or copyright law. Dissemination of this information or
reproduction of this material is strictly forbidden unless prior written
permission is obtained from {{.Owner}}
`

	readmeTemplate = `# {{upper .ServiceName}}

## Prepare service

` + "```" + `
$ go mod tidy
` + "```" + `

### If include GRPC handler, run this command (must install ` + "`protoc`" + ` compiler min version ` + "`libprotoc` 3.14.0`" + `)

` + "```" + `
$ make proto
` + "```" + `

### If using SQL database, run this commands for migration

Create new migration:
` + "```" + `
$ make migration create [your_migration_name]
` + "```" + `

UP migration:
` + "```" + `
$ make migration
` + "```" + `

Rollback migration:
` + "```" + `
$ make migration down
` + "```" + `

## Build and run service
` + "```" + `
$ make run
` + "```" + `

## Run unit test & calculate code coverage

Make sure generate mock using [mockery](https://github.com/vektra/mockery)
` + "```" + `
$ make mocks
` + "```" + `

Run test:
` + "```" + `
$ make test
` + "```" + `

## Create docker image
` + "```" + `
$ make docker
` + "```"

	readmeMonorepoTemplate = "# Backend Microservices\n\n" +
		"## Build with\n" +
		`<p align="center">` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/golang.png" width="80" alt="golang logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/docker.png" width="80" hspace="10" alt="docker logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/rest.png" width="80" hspace="10" alt="rest logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/graphql.png" width="80" alt="graphql logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/grpc.png" width="160" hspace="15" vspace="15" alt="grpc logo" />` + "\n" +
		` <img src="https://storage.googleapis.com/agungdp/static/logo/kafka.png" height="80" alt="kafka logo" />` + "\n" +
		"</p>\n\n" +
		"This repository explain implementation of Go for building multiple microservices using a single codebase. Using [Standard Golang Project Layout](https://github.com/golang-standards/project-layout) and [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)\n\n" +
		"## Create new service (for new project)\n" +
		"Please install **latest** [**candi**](https://github.com/golangid/candi) CLI first, and then:\n" +
		"```\n" +
		"$ candi -init\n" +
		"```\n" +
		`
### If include GRPC handler, run this command (must install ` + "`protoc`" + ` compiler min version ` + "`libprotoc` 3.14.0`" + `)

` + "```" + `
$ make proto service={{service_name}}
` + "```" + `

### If using SQL database, run this commands for migration

Create new migration:
` + "```" + `
$ make migration service={{service_name}} create [your_migration_name]
` + "```" + `

UP migration:
` + "```" + `
$ make migration service={{service_name}}
` + "```" + `

Rollback migration:
` + "```" + `
$ make migration service={{service_name}} down
` + "```\n\n" +
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
		"* **Make sure generate mock using [mockery](https://github.com/vektra/mockery):**\n" +
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
FROM golang:1.18.4-alpine3.16 AS dependency_builder

WORKDIR /go/src
ENV GO111MODULE=on

RUN apk update
RUN apk add --no-cache bash ca-certificates git make gcc musl-dev

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
RUN go build -o bin services/$SERVICE_NAME/*.go

# Stage 3
FROM alpine:latest  

ARG BUILD_NUMBER
ARG SERVICE_NAME
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /usr/app/
ENV WORKDIR=services/$SERVICE_NAME/
ENV BUILD_NUMBER=$BUILD_NUMBER

RUN mkdir -p /usr/app/services/$SERVICE_NAME
RUN mkdir -p /usr/app/services/$SERVICE_NAME/api
COPY --from=service_builder /usr/app/bin bin
COPY --from=service_builder /usr/app/services/$SERVICE_NAME/.env /usr/app/services/$SERVICE_NAME/.env
COPY --from=service_builder /usr/app/services/$SERVICE_NAME/api /usr/app/services/$SERVICE_NAME/api

ENTRYPOINT ["./bin"]
`

	makefileMonorepoTemplate = `.PHONY : build run

args = ` + "`arg=\"$(filter-out $@,$(MAKECMDGOALS))\" && echo $${arg:-${1}}`" + `

$(eval $(service):;@:)

check:
	@[ "${service}" ] || ( echo "\x1b[31;1mERROR: 'service' is not set\x1b[0m"; exit 1 )
	@if [ ! -d "services/$(service)" ]; then  echo "\x1b[31;1mERROR: service '$(service)' undefined\x1b[0m"; exit 1; fi

prepare: check
	@if [ ! -f services/$(service)/.env ]; then cp services/$(service)/.env.sample services/$(service)/.env; fi;

init:
	@candi --init

add-module: check
	@candi --add-module --service=$(service)

proto: check
	@if [ ! -d "sdk/$(service)/proto" ]; then echo "creating new proto files..." && mkdir sdk/$(service)/proto; fi
	$(foreach proto_file, $(shell find services/$(service)/api/proto -name '*.proto'),\
	protoc --proto_path=services/$(service)/api/proto --go_out=plugins=grpc:sdk/$(service)/proto \
	--go_opt=paths=source_relative $(proto_file);)

migration: check
	@WORKDIR="services/$(service)/" go run services/$(service)/cmd/migration/migration.go $(call args,up)

build: check
	@go build -o services/$(service)/bin services/$(service)/*.go

run: build
	@WORKDIR="services/$(service)/" ./services/$(service)/bin

docker: check
	docker build --build-arg SERVICE_NAME=$(service) -t $(service):latest .

run-container: check
	docker run --name=$(service) --network="host" -d $(service)

clear-docker: check
	docker rm -f $(service)
	docker rmi -f $(service)

# mocks all interfaces from selected service for unit test
mocks: check
	@mockery --all --keeptree --dir=sdk --output=./sdk/mocks
	@if [ -f sdk/mocks/Option.go ]; then rm sdk/mocks/Option.go; fi;
	@mockery --all --keeptree --dir=globalshared --output=./globalshared/mocks
	@mockery --all --keeptree --dir=services/$(service)/internal --output=services/$(service)/pkg/mocks --case underscore
	@mockery --all --keeptree --dir=services/$(service)/pkg --output=services/$(service)/pkg/mocks --case underscore

# unit test & calculate code coverage from selected service (please run mocks before run this rule)
test: check
	@echo "\x1b[32;1m>>> running unit test and calculate coverage for service $(service)\x1b[0m"
	@if [ -f services/$(service)/coverage.txt ]; then rm services/$(service)/coverage.txt; fi;
	@go test ./services/$(service)/... -cover -coverprofile=services/$(service)/coverage.txt -covermode=count \
		-coverpkg=$$(go list ./services/$(service)/... | grep -v mocks | tr '\n' ',')
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
