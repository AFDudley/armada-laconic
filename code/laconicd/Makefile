#!/usr/bin/make -f

DOCKER := $(shell which docker)

LACONIC_BINARY = laconicd

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
	VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

# Update the ldflags with the app, client & server names
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=laconic \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(LACONIC_BINARY) \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

BUILDDIR ?= $(CURDIR)/build

###########
# Install #
###########

go.sum: go.mod
	@echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/
build-linux:
	GOOS=linux GOARCH=amd64 LEDGER_ENABLED=false $(MAKE) build

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	@echo "--> $@ing laconicd"
	go $@ $(BUILD_FLAGS) $(BUILD_ARGS) ./cmd/laconicd

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

all: install

.PHONY: build build-linux install

##################
###  Protobuf  ###
##################

protoVer=0.17.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoVolumes=-v $(shell readlink -f ..):/mod
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoVolumes) $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating protobuf files..."
	@$(protoRun) sh ./scripts/protocgen.sh
	@go mod tidy

proto-format:
	@$(protoRun) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoRun) buf lint proto/ --error-format=json

.PHONY: proto-all proto-gen proto-format proto-lint

#################
###  Linting  ###
#################

golangci_lint_cmd=golangci-lint
golangci_version=v1.51.2

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run ./... --timeout 15m

lint-fix:
	@echo "--> Running linter and fixing issues"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run ./... --fix --timeout 15m

.PHONY: lint lint-fix

#################
###  Tests  ###
#################

test-integration:
	$(MAKE) -C tests test-integration

test-e2e:
	$(MAKE) -C tests test-e2e

test-unit:
	go test ./utils/... ./cmd/... -mod=readonly -test.v

test-system: build
	test -d ./tests/system/binaries || ln -sf $(BUILDDIR) ./tests/system/binaries
	go test ./tests/system -v
