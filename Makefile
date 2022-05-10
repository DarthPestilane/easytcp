export CGO_ENABLED=0

default: build
ldflags=-ldflags="-s"
build:
	go build ${ldflags} -v

build-all:
	go build ${ldflags} -v ./...

lint:
	golangci-lint run --concurrency=2

lint-fix:
	golangci-lint run --concurrency=2 --fix

test:
	CGO_ENABLED=1 go test -count=1 -race -covermode=atomic -coverprofile=.testCoverage.txt -timeout=2m .

test-v:
	CGO_ENABLED=1 go test -count=1 -race -covermode=atomic -coverprofile=.testCoverage.txt -timeout=2m -v .

cover-view:
	go tool cover -func .testCoverage.txt
	go tool cover -html .testCoverage.txt

check: test lint
	go tool cover -func .testCoverage.txt

bench:
	go test -bench=. -run=none -benchmem -benchtime=250000x

tidy:
	go mod tidy -v

os=`uname`
gen:
ifeq (${os}, $(filter ${os}, Windows Windows_NT)) # If on windows, there might be something unexpected.
	rm -rf ./**/gomock_reflect_*
	go generate 2>/dev/null
	rm -rf ./**/gomock_reflect_*
else
	go generate -v
endif

release-local:
	goreleaser release --rm-dist --skip-announce --skip-publish --snapshot

clean:
	go clean -r -x -cache -i
