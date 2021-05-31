.PHONY: *

ldflags = -ldflags="-s -w"

build: # build
	CGO_ENABLED=0 go build ${ldflags} -v ./...

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${ldflags} -v ./...

install:
	CGO_ENABLED=0 go get -v -insecure -t -d

lint: # 代码风格检查
	CGO_ENABLED=0 golangci-lint run --concurrency=2

lint-fix:
	CGO_ENABLED=0 golangci-lint run --concurrency=2 --fix

test:
	CGO_ENABLED=0 go test -count=1 ./...

spec: lint test # 语法检查+单元测试

tidy:
	go mod tidy
