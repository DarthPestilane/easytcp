ldflags = -ldflags="-s -w"
coverprofile=.testCoverage.txt
pkgs=`go list ./... | grep -v /examples/`

.PHONY: default
default: build

.PHONY: build
build:
	CGO_ENABLED=0 go build ${ldflags} -v ${pkgs}

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
	CGO_ENABLED=0 go test -count=1 -covermode=set -coverprofile=${coverprofile} ${pkgs}

.PHONY: coverage
coverage:
	CGO_ENABLED=0 go tool cover -html ${coverprofile}

.PHONY: spec
spec: lint test

.PHONY: bench
bench:
	CGO_ENABLED=0 go test -bench=. -run=none -benchmem ${pkgs}

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: gen
gen:
ifeq ($(OS), $(filter $(OS), Windows Windows_NT)) # If on windows, there might be something strange...
	rm -rf ./**/gomock_reflect_*
	CGO_ENABLED=0 go generate ${pkgs} 2>/dev/null
	rm -rf ./**/gomock_reflect_*
else
	CGO_ENABLED=0 go generate ${pkgs}
endif
