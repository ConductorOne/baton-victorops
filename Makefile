GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
BUILD_DIR = dist/${GOOS}_${GOARCH}
GENERATED_CONF = pkg/config/conf.gen.go

ifeq ($(GOOS),windows)
OUTPUT_PATH = ${BUILD_DIR}/baton-victorops.exe
else
OUTPUT_PATH = ${BUILD_DIR}/baton-victorops
endif

ifdef BATON_LAMBDA_SUPPORT
BUILD_TAGS = -tags baton_lambda_support
endif

.PHONY: build
build: $(GENERATED_CONF)
	go build ${BUILD_TAGS} -o ${OUTPUT_PATH} ./cmd/baton-victorops

$(GENERATED_CONF): pkg/config/config.go go.mod
	go generate ./pkg/config

.PHONY: generate
generate:
	go generate ./pkg/config

.PHONY: update-deps
update-deps:
	go get -d -u ./...
	go mod tidy -v
	go mod vendor

.PHONY: add-dep
add-dep:
	go mod tidy -v
	go mod vendor

.PHONY: lint
lint:
	golangci-lint run
