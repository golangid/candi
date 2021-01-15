package main

const (
	dockerfileTemplate = `# Stage 1
FROM golang:1.14.9-alpine3.12 AS dependency_builder

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

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

RUN mkdir -p /root/api
RUN mkdir -p /root/cmd/{{.ServiceName}}
RUN mkdir -p /root/config/key
COPY --from=service_builder /usr/app/bin bin
COPY --from=service_builder /usr/app/.env .env
COPY --from=service_builder /usr/app/api /root/api

ENTRYPOINT ["./bin"]
`

	makefileTemplate = `.PHONY : build run

build:
	go build -o bin

run: build
	./bin

proto:
	$(foreach proto_file, $(shell find api/proto -name '*.proto'),\
	protoc --proto_path=api/proto --go_out=plugins=grpc:api/proto \
	--go_opt=paths=source_relative $(proto_file);)

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

go 1.14

require pkg.agungdwiprasetyo.com/candi {{.Version}}

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
)
