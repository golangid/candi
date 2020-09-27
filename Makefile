.PHONY : test

PACKAGES = $(shell go list ./... | grep -v -e . -e mocks | tr '\n' ',')

# unit test & calculate code coverage
test:
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo ">> running unit test and calculate coverage"
	@go test ./... -cover -coverprofile=coverage.txt -covermode=count -coverpkg=$(PACKAGES)
	@go tool cover -func=coverage.txt
