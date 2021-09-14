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
	CGO_ENABLED=1 go test -race -count=1 -covermode=atomic -coverprofile=.testCoverage.txt -timeout=2m . ./message

.PHONY: test-v
test-v:
	CGO_ENABLED=1 go test -race -count=1 -covermode=atomic -coverprofile=.testCoverage.txt -timeout=2m -v . ./message

.PHONY: cover-view
cover-view:
	go tool cover -func .testCoverage.txt
	go tool cover -html .testCoverage.txt

.PHONY: spec
spec: lint test
	go tool cover -func .testCoverage.txt

.PHONY: bench
bench:
	go test -bench=. -run=none -benchmem -benchtime=250000x

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: gen
gen:
ifeq (${os}, $(filter ${os}, Windows Windows_NT)) # If on windows, there might be something unexpected.
	rm -rf ./**/gomock_reflect_*
	go generate 2>/dev/null
	rm -rf ./**/gomock_reflect_*
else
	go generate -v
endif

.PHONY: release-local
release-local:
	goreleaser release --rm-dist --skip-announce --skip-publish --snapshot

.PHONY: clean
clean:
	go clean -r -x -cache -i
