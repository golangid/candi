# Stage 1
FROM golang:1.12.7-alpine3.10 AS dependency_builder

RUN apk add bash ca-certificates git make gcc g++ libc-dev
WORKDIR /go/src
ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

# Stage 2
FROM dependency_builder AS service_builder

ARG SERVICE_NAME
WORKDIR /usr/app

COPY . .
RUN make prepare service=$SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -a -o bin

# Stage 3
FROM alpine:latest  

ARG SERVICE_NAME
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

RUN mkdir -p /root/api/$SERVICE_NAME
RUN mkdir -p /root/cmd/$SERVICE_NAME
RUN mkdir -p /root/config/key
COPY --from=service_builder /usr/app/bin bin
COPY --from=service_builder /usr/app/cmd/$SERVICE_NAME/.env /root/cmd/$SERVICE_NAME/.env
COPY --from=service_builder /usr/app/api/$SERVICE_NAME /root/api/$SERVICE_NAME
COPY --from=service_builder /usr/app/config/key /root/config/key

ENTRYPOINT ["./bin"]
