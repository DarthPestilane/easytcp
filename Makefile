.PHONY: *

ldflags = -ldflags="-s -w"

default: build

build:
	CGO_ENABLED=0 go build ${ldflags} -v `go list ./... | grep -v /examples/`

build-all:
	CGO_ENABLED=0 go build ${ldflags} -v ./...

lint:
	CGO_ENABLED=0 golangci-lint run --concurrency=2

lint-fix:
	CGO_ENABLED=0 golangci-lint run --concurrency=2 --fix

coverprofile=.testCoverage.txt

test:
	CGO_ENABLED=0 go test -count=1 -covermode=atomic -coverprofile=${coverprofile} `go list ./... | grep -v /examples/`

coverage:
	CGO_ENABLED=0 go tool cover -html ${coverprofile}

spec: lint test

tidy:
	go mod tidy

gen:
	CGO_ENABLED=0 go generate `go list ./... | grep -v /examples/` &>/dev/null
	rm -rf ./**/gomock_reflect_*
