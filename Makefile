.PHONY: all
export GO111MODULE=on

all: clean build test

APP=dobby-service
APP_VERSION:=$(shell cat .version)
APP_COMMIT:=$(shell git rev-parse HEAD)
APP_EXECUTABLE="./out/$(APP)"
ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")
UNAME := $(shell uname -m)
GO_FLAGS_x86_64 :=
GO_FLAGS_arm64 := -tags dynamic

assign-vars = $(if $(1),$(1),$(shell grep '$(2):' application.yml | tail -n1| cut -d':' -f2))

gen-wire-deps:
	cd cmd/dobby-service && wire && cd ../..

compile:
	mkdir -p out/
	go build -o $(APP_EXECUTABLE) -ldflags "-X main.version=$(APP_VERSION) -X main.commit=$(APP_COMMIT)" $(GO_FLAGS_$(UNAME)) ./cmd/dobby-service

static-check:
	go install honnef.co/go/tools/cmd/staticcheck@v0.4.7
	staticcheck $(GO_FLAGS_$(UNAME)) $(ALL_PACKAGES)

fmt:
	go fmt $(ALL_PACKAGES)

vet:
	go vet $(GO_FLAGS_$(UNAME)) $(ALL_PACKAGES)

lint:
	@for p in $(ALL_PACKAGES); do \
		echo "==> Linting $$p"; \
		golint $$p | { grep -vwE "exported (var|function|method|type|const) \S+ should have comment" || true; } \
	done

build: compile fmt vet lint static-check

clean:
	rm -rf out/

test: set-packages-to-test test-cover-html

test-cover:
	go get -u github.com/jokeyrhyme/go-coverage-threshold/cmd/go-coverage-threshold
	ENVIRONMENT=test go-coverage-threshold

set-packages-to-test:
	$(eval PACKAGES_TO_TEST := $(shell go list ./... | grep -v "internal/mongo\|internal/redis\|internal/testutils"))

test-cover-html:
	mkdir -p out/
	go test -covermode=count  -coverprofile=cover.out.tmp ${PACKAGES_TO_TEST}
	cat cover.out.tmp | grep -v "_mock.go" | grep -v "utils/time.go\|health/service.go" > coverage-all.out
	rm cover.out.tmp
	@go tool cover -html=coverage-all.out -o out/coverage.html
	@go tool cover -func=coverage-all.out

ci: clean build test

build-run-server: build
	${APP_EXECUTABLE} start
