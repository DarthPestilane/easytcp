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

coverprofile=.testCoverage.txt

test:
	CGO_ENABLED=0 go test -count=1 -cover -coverprofile=${coverprofile} `go list ./... | grep -v /examples/`

coverage:
	CGO_ENABLED=0 go tool cover -html ${coverprofile}

spec: lint test

tidy:
	go mod tidy

gen:
	CGO_ENABLED=0 go generate `go list ./... | grep -v /examples/` &>/dev/null
	rm -rf ./**/gomock_reflect_*
