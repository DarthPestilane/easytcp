.PHONY: *

ldflags = -ldflags="-s -w"

build:
	CGO_ENABLED=0 go build ${ldflags} -v ./...

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${ldflags} -v ./...

install:
	CGO_ENABLED=0 go get -v -insecure -t -d

lint:
	CGO_ENABLED=0 golangci-lint run --concurrency=2

lint-fix:
	CGO_ENABLED=0 golangci-lint run --concurrency=2 --fix

test:
	CGO_ENABLED=0 go test -count=1 -cover -coverprofile=.testCoverage.txt `go list ./... | grep -v /examples/`

coverage:
	CGO_ENABLED=0 go tool cover -html .testCoverage.txt

spec: lint test

tidy:
	go mod tidy
