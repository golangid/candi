.PHONY : mocks test

# mocks all interfaces
mocks:
	@mockery --all --keeptree
	@if [ -f mocks/candiutils/HTTPRequestOption.go ]; then rm mocks/candiutils/HTTPRequestOption.go; fi;
	@if [ -f mocks/candiutils/httpClientDo.go ]; then rm mocks/candiutils/httpClientDo.go; fi;
	@if [ -f mocks/codebase/app/cron_worker/OptionFunc.go ]; then rm mocks/codebase/app/cron_worker/OptionFunc.go; fi;
	@if [ -f mocks/codebase/app/graphql_server/ws/sendFunc.go ]; then rm mocks/codebase/app/graphql_server/ws/sendFunc.go; fi;
	@if [ -f mocks/codebase/app/graphql_server/OptionFunc.go ]; then rm mocks/codebase/app/graphql_server/OptionFunc.go; fi;
	@if [ -f mocks/codebase/app/grpc_server/OptionFunc.go ]; then rm mocks/codebase/app/grpc_server/OptionFunc.go; fi;
	@if [ -f mocks/codebase/app/kafka_worker/OptionFunc.go ]; then rm mocks/codebase/app/kafka_worker/OptionFunc.go; fi;
	@if [ -f mocks/codebase/app/postgres_worker/OptionFunc.go ]; then rm mocks/codebase/app/postgres_worker/OptionFunc.go; fi;
	@if [ -f mocks/codebase/app/rabbitmq_worker/OptionFunc.go ]; then rm mocks/codebase/app/rabbitmq_worker/OptionFunc.go; fi;
	@if [ -f mocks/codebase/app/redis_worker/OptionFunc.go ]; then rm mocks/codebase/app/redis_worker/OptionFunc.go; fi;
	@if [ -f mocks/codebase/app/rest_server/OptionFunc.go ]; then rm mocks/codebase/app/rest_server/OptionFunc.go; fi;
	@if [ -f mocks/codebase/app/task_queue_worker/OptionFunc.go ]; then rm mocks/codebase/app/task_queue_worker/OptionFunc.go; fi;
	@if [ -f mocks/codebase/factory/dependency/Option.go ]; then rm mocks/codebase/factory/dependency/Option.go; fi;
	@if [ -f mocks/validator/JSONSchemaValidatorOptionFunc.go ]; then rm mocks/validator/JSONSchemaValidatorOptionFunc.go; fi;
	@if [ -f mocks/validator/StructValidatorOptionFunc.go ]; then rm mocks/validator/StructValidatorOptionFunc.go; fi;
	@rm -rf mocks/cmd;

# unit test & calculate code coverage
test:
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo ">> running unit test and calculate coverage"
	@go test -race ./... -cover -coverprofile=coverage.txt -covermode=atomic \
		-coverpkg=$$(go list ./... | grep -v -e mocks -e codebase -e cmd | tr '\n' ',')
	@go tool cover -func=coverage.txt
