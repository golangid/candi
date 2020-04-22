.PHONY : prepare build run

SERVICE_NAME := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(SERVICE_NAME):;@:)
ifndef SERVICE_NAME
$(error SERVICE_NAME is not set)
endif

PROTO_FILES := $(shell find api/proto/$(SERVICE_NAME) -name '*.proto')

prepare:
	ln -sf cmd/$(SERVICE_NAME)/main.go main_service.go
	$(foreach proto_file, $(PROTO_FILES),\
	protoc -I . $(proto_file) --go_out=plugins=grpc:.;)

build: prepare
	go build -o bin

run: build
	./bin

docker: prepare
	docker build --build-arg SERVICE_NAME=$(SERVICE_NAME) -t $(SERVICE_NAME):latest .

run-container:
	docker run --name=$(SERVICE_NAME) --network="host" -d $(SERVICE_NAME)

clear:
	rm main_service.go bin backend-microservices warung wedding
