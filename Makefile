.PHONY : mocks test

# mocks all interfaces
mocks:
	@mockery --all --keeptree
	@rm -rf mocks/candiutils/HTTPRequestOption.go
	@rm -rf mocks/candiutils/httpClientDo.go
	@rm -rf mocks/codebase/app/cron_worker/OptionFunc.go
	@rm -rf mocks/codebase/app/graphql_server/ws/sendFunc.go
	@rm -rf mocks/codebase/app/graphql_server/OptionFunc.go
	@rm -rf mocks/codebase/app/grpc_server/OptionFunc.go
	@rm -rf mocks/codebase/app/kafka_worker/OptionFunc.go
	@rm -rf mocks/codebase/app/postgres_worker/OptionFunc.go
	@rm -rf mocks/codebase/app/rabbitmq_worker/OptionFunc.go
	@rm -rf mocks/codebase/app/redis_worker/OptionFunc.go
	@rm -rf mocks/codebase/app/rest_server/OptionFunc.go
	@rm -rf mocks/codebase/app/task_queue_worker/OptionFunc.go
	@rm -rf mocks/codebase/factory/dependency/Option.go
	@rm -rf mocks/validator/JSONSchemaValidatorOptionFunc.go
	@rm -rf mocks/validator/StructValidatorOptionFunc.go
	@rm -rf mocks/candishared/DBUpdateOptionFunc.go;
	@rm -rf mocks/logger/Masker.go;
	@rm -rf mocks/config;
	@rm -rf mocks/cmd;

# unit test & calculate code coverage
test:
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo ">> running unit test and calculate coverage"
	@go test -race ./... -cover -coverprofile=coverage.txt -covermode=atomic \
		-coverpkg=$$(go list ./... | grep -v -e mocks -e codebase -e cmd | tr '\n' ',')
	@go tool cover -func=coverage.txt
