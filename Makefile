.PHONY : prepare build run

SERVICE_NAME := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(SERVICE_NAME):;@:)

prepare:
	if [ -f main.go ]; then rm main.go; fi;
	ln -s cmd/$(SERVICE_NAME)/main.go main.go

build: prepare
	go build -o bin

run: build
	./bin

docker: prepare
	docker build --build-arg SERVICE_NAME=$(SERVICE_NAME) -t $(SERVICE_NAME):latest .

run-container:
	docker run --name=$(SERVICE_NAME) --network="host" $(SERVICE_NAME)

clear:
	if [ -f main.go ]; then echo "ADA"; fi;
	rm main.go bin backend-microservices warung wedding