# Run `make help` to display help

# --- Global -------------------------------------------------------------------
O = out
COVERAGE = 64
VERSION ?= $(shell git describe --tags --dirty  --always)
REPO_ROOT = $(shell git rev-parse --show-toplevel)

all: build test check-coverage lint  ## build, test, check coverage and lint
	@if [ -e .git/rebase-merge ]; then git --no-pager log -1 --pretty='%h %s'; fi
	@echo '$(COLOUR_GREEN)Success$(COLOUR_NORMAL)'

ci: clean ci-protos all  ## Full clean build and up-to-date checks as run on CI

clean::  ## Remove generated files
	-rm -rf $(O)

.PHONY: all ci clean

# --- Build --------------------------------------------------------------------
GO_LDFLAGS = -X main.version=$(VERSION)

build: | $(O)  ## Build reflect binaries
	go build -o $(O)/reflect -ldflags='$(GO_LDFLAGS)' .
	go build -o $(O)/testserver ./cmd/testserver

install:  ## Build and install binaries in $GOBIN
	go install -ldflags='$(GO_LDFLAGS)' .

run: build  ## Run test server
	$(O)/testserver

.PHONY: build install

# --- Test ---------------------------------------------------------------------
COVERFILE = $(O)/coverage.txt

test: ## Run tests and generate a coverage file
	go test -coverprofile=$(COVERFILE) ./...

check-coverage: test  ## Check that test coverage meets the required level
	@go tool cover -func=$(COVERFILE) | $(CHECK_COVERAGE) || $(FAIL_COVERAGE)

cover: test  ## Show test coverage in your browser
	go tool cover -html=$(COVERFILE)

CHECK_COVERAGE = awk -F '[ \t%]+' '/^total:/ {print; if ($$3 < $(COVERAGE)) exit 1}'
FAIL_COVERAGE = { echo '$(COLOUR_RED)FAIL - Coverage below $(COVERAGE)%$(COLOUR_NORMAL)'; exit 1; }

.PHONY: build-test check-coverage cover test test-short

# --- Lint ---------------------------------------------------------------------

lint:  ## Lint go source code
	golangci-lint run

.PHONY: lint

# --- Protos -------------------------------------------------------------------
PROTO_DIR = protos
PROTO_VENDOR_DIR = $(PROTO_DIR)/vendor
PROTO_FILES = $(shell find $(PROTO_DIR) -path $(PROTO_VENDOR_DIR) -prune -o -name '*.proto' -print)
PKG_GEN_DIRS = $(sort $(patsubst $(PROTO_DIR)/%,pkg/%,$(dir $(PROTO_FILES))))
PROTOC_GO_FLAGS = \
	-I $(PROTO_DIR) \
	-I $(PROTO_VENDOR_DIR) \
	--go_out=paths=source_relative:pkg \
	--go-grpc_out=paths=source_relative:pkg

ci-protos: install-proto-tools vendor-protos check-protos

protos: protos-echo3
	@goimports -w $(PKG_GEN_DIRS)

protos-%:  ## Generate go files from proto and gRPC definitions
	protoc $(PROTOC_GO_FLAGS) protos/$*/$*.proto

check-protos: protos  ## Check that generated proto and gRPC code is up-to-date
	test -z "$$(git status --porcelain $(PKG_GEN_DIRS))"

install-proto-tools:  ## Install protoc plugins to generate go code from proto and gRPC definitions
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.25.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.15.2
	go install golang.org/x/tools/cmd/goimports@v0.1.0

vendor-api-%:
	curl -fsSL --create-dirs -o $(PROTO_VENDOR_DIR)/google/api/$*.proto https://github.com/googleapis/googleapis/raw/master/google/api/$*.proto

vendor-pb-%:
	curl -fsSL --create-dirs -o $(PROTO_VENDOR_DIR)/google/protobuf/$*.proto https://github.com/protocolbuffers/protobuf/raw/master/src/google/protobuf/$*.proto

vendor-protos: vendor-api-annotations vendor-api-http vendor-pb-descriptor vendor-pb-any

clean::
	rm -rf $(addsuffix *.pb.go,$(PKG_GEN_DIRS))
	rm -rf $(addsuffix *.pb.gw.go,$(PKG_GEN_DIRS))
	rm -rf $(PROTO_VENDOR_DIR)

.PHONY: check-protos ci-protos install-proto-tools protos vendor-protos

# --- Release -------------------------------------------------------------------
NEXTTAG := $(shell { git tag --list --merged HEAD --sort=-v:refname; echo v0.0.0; } | grep -E "^v?[0-9]+.[0-9]+.[0-9]+$$" | head -n1 | awk -F . '{ print $$1 "." $$2 "." $$3 + 1 }')

release: releasetag  ## Tag and release binaries for different OS on GitHub release
	goreleaser release --rm-dist

releasetag:
	git tag $(NEXTTAG)

.PHONY: dist publish-s3 release releasetag

# --- Utilities ----------------------------------------------------------------
COLOUR_NORMAL = $(shell tput sgr0 2>/dev/null)
COLOUR_RED    = $(shell tput setaf 1 2>/dev/null)
COLOUR_GREEN  = $(shell tput setaf 2 2>/dev/null)
COLOUR_WHITE  = $(shell tput setaf 7 2>/dev/null)

help:
	@awk -F ':.*## ' 'NF == 2 && $$1 ~ /^[A-Za-z0-9%_-]+$$/ { printf "$(COLOUR_WHITE)%-25s$(COLOUR_NORMAL)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

$(O):
	@mkdir -p $@

.PHONY: help
