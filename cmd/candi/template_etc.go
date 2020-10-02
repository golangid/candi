package main

const (
	dockerfileTemplate = `# Stage 1
FROM golang:1.12.7-alpine3.10 AS dependency_builder

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
RUN make prepare
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -o bin

# Stage 3
FROM alpine:latest  

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

RUN mkdir -p /root/api
RUN mkdir -p /root/cmd/{{.ServiceName}}
RUN mkdir -p /root/config/key
COPY --from=service_builder /usr/app/bin bin
COPY --from=service_builder /usr/app/cmd/{{.ServiceName}}/.env /root/cmd/{{.ServiceName}}/.env
COPY --from=service_builder /usr/app/api /root/api

ENTRYPOINT ["./bin"]
`

	makefileTemplate = `.PHONY : prepare build run

build: prepare
	go build -o bin

run: build
	./bin

proto:
	$(foreach proto_file, $(shell find api/proto -name '*.proto'),\
	protoc -I . $(proto_file) --go_out=plugins=grpc:.;)

docker: prepare
	docker build -t {{.ServiceName}}:latest .

run-container:
	docker run --name={{.ServiceName}} --network="host" -d {{.ServiceName}}

clear:
	rm main_service.go bin {{.ServiceName}}
`

	gomodTemplate = `module {{.ServiceName}}

go 1.14
`

	gitignoreTemplate = `bin
vendor
main_service.go
`
)
