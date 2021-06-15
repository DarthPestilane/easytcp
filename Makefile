ldflags = -ldflags="-s -w"
coverprofile=.testCoverage.txt

.PHONY: default
default: build

.PHONY: build
build:
	CGO_ENABLED=0 go build ${ldflags} -v `go list ./... | grep -v /examples/`

.PHONY: build-all
build-all:
	CGO_ENABLED=0 go build ${ldflags} -v ./...

.PHONY: lint
lint:
	CGO_ENABLED=0 golangci-lint run --concurrency=2

.PHONY: lint-fix
lint-fix:
	CGO_ENABLED=0 golangci-lint run --concurrency=2 --fix

.PHONY: test
test:
	CGO_ENABLED=0 go test -count=1 -covermode=set -coverprofile=${coverprofile} `go list ./... | grep -v /examples/`

.PHONY: coverage
coverage:
	CGO_ENABLED=0 go tool cover -html ${coverprofile}

.PHONY: spec
spec: lint test

.PHONY: bench
bench:
	CGO_ENABLED=0 go test -bench=. -run=none -benchmem `go list ./... | grep -v /examples/`

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: gen
gen:
	CGO_ENABLED=0 go generate `go list ./... | grep -v /examples/` &>/dev/null
	rm -rf ./**/gomock_reflect_*
