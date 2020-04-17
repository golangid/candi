.PHONY : prepare build run

SERVICE_NAME := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(SERVICE_NAME):;@:)
ifndef SERVICE_NAME
$(error SERVICE_NAME is not set)
endif

prepare:
	ln -sf cmd/$(SERVICE_NAME)/main.go main_service.go

build: prepare
	go build -o bin

run: build
	./bin

docker: prepare
	docker build --build-arg SERVICE_NAME=$(SERVICE_NAME) -t $(SERVICE_NAME):latest .

run-container:
	docker run --name=$(SERVICE_NAME) --network="host" -d $(SERVICE_NAME)

clear:
	if [ -f main.go ]; then echo "ADA"; fi;
	rm main.go bin backend-microservices warung wedding