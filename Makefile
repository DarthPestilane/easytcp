ldflags=-ldflags="-s"
os=`uname`

export CGO_ENABLED=0

.PHONY: default
default: build

.PHONY: build
build:
	go build ${ldflags} -v

.PHONY: build-all
build-all:
	go build ${ldflags} -v ./...

.PHONY: lint
lint:
	golangci-lint run --concurrency=2

.PHONY: lint-fix
lint-fix:
	golangci-lint run --concurrency=2 --fix

.PHONY: test
test:
	go test -count=1 -covermode=set -coverprofile=.testCoverage.txt -timeout=2m .

.PHONY: test-v
test-v:
	go test -count=1 -covermode=set -coverprofile=.testCoverage.txt -timeout=2m -v .

.PHONY: cover-view
cover-view:
	go tool cover -func .testCoverage.txt
	go tool cover -html .testCoverage.txt

.PHONY: spec
spec: lint test
	go tool cover -func .testCoverage.txt

.PHONY: bench
bench:
	go test -bench=. -run=none -benchmem

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: gen
gen:
ifeq (${os}, $(filter ${os}, Windows Windows_NT)) # If on windows, there might be something unexpected.
	rm -rf ./**/gomock_reflect_*
	CGO_ENABLED=0 go generate 2>/dev/null
	rm -rf ./**/gomock_reflect_*
else
	CGO_ENABLED=0 go generate
endif
