PROJECT_NAME := ocm-container

# Container engine detection
CONTAINER_ENGINE := $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null)
ifeq ($(CONTAINER_ENGINE),)
$(error ERROR: container engine not found. Install podman or docker and run make again)
endif

REGISTRY_USER         ?= $(QUAY_USER)
REGISTRY_TOKEN        ?= $(QUAY_TOKEN)

IMAGE_REGISTRY        ?= quay.io
IMAGE_REPOSITORY      ?= app-sre
IMAGE_NAME            ?= $(PROJECT_NAME)
IMAGE_URI             := $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)
TAG                   ?= latest
GIT_REVISION          := $(shell git rev-parse --short=7 HEAD)

BUILD_ARGS            ?=
CACHE                 ?= --no-cache

ifdef GITHUB_TOKEN
GITHUB_BUILD_ARGS     := --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)
endif

# Architecture detection
RAW_ARCHITECTURE ?= $(shell arch)
ARCHITECTURE     := $(patsubst aarch64,arm64,$(patsubst x86_64,amd64,$(RAW_ARCHITECTURE)))

# Golang build settings
unexport GOFLAGS
GOOS     ?= linux
GOARCH   ?= $(ARCHITECTURE)
GOENV     = GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 GOFLAGS=
GOPATH   := $(shell go env GOPATH)
HOME     ?= $(shell mktemp -d)
TESTOPTS ?=

export GO111MODULE = on
export GOPROXY     = https://proxy.golang.org
export CGO_ENABLED = 0

# Tool configs
GOLANGCI_LINT_VERSION      := v2.1.6
GORELEASER_VERSION         := v2.43.0
GORELEASER_CONFIG          := .goreleaser.yaml
GORELEASER_CORES           := 4
GORELEASER_ADDITIONAL_ARGS ?=

# Helper macro: $(call build_target,<image name>,<architecture>
define build_target
	$(CONTAINER_ENGINE) build $(CACHE) $(BUILD_ARGS) ${GITHUB_BUILD_ARGS} -f Containerfile --platform=linux/$(2) --target=$(1) -t $(1):$(2) .
endef

# Helper macro: $(call tag_target,<image name>, <build id>)
define tag_target
	${CONTAINER_ENGINE} tag $(1):$(ARCHITECTURE) $(IMAGE_URI)/$(1):$(2)-$(GIT_REVISION)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} tag $(1):$(ARCHITECTURE) $(IMAGE_URI)/$(1):$(2)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} tag $(1):$(ARCHITECTURE) $(IMAGE_URI)/$(1):latest-$(ARCHITECTURE)
endef

# Helper macro: $(call push_target,<image name>,<build id>)
define push_target
	${CONTAINER_ENGINE} push $(IMAGE_URI)/$(1):$(2)-$(GIT_REVISION)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} push $(IMAGE_URI)/$(1):$(2)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} push $(IMAGE_URI)/$(1):latest-$(ARCHITECTURE)
endef

# Helper macro: $(call get_build_id,<image name>,<architecture>)
# The build ID is the short hash of the image, which is used for tagging
# This retrieves the build ID of a specific image and architecture
define get_build_id
	$(shell ${CONTAINER_ENGINE} image inspect $(1):$(2) | jq -r '.[].Id' | cut -c 1-12)
endef

# Build targets
.PHONY: build-all build-micro build-minimal build-full build
build-all: build-micro build-minimal build-full

build-micro:
	@$(call build_target,$(IMAGE_NAME)-micro,$(ARCHITECTURE))

build-minimal:
	@$(call build_target,$(IMAGE_NAME)-minimal,$(ARCHITECTURE))

build-full:
	@$(call build_target,$(IMAGE_NAME),$(ARCHITECTURE))

build: build-full

.PHONY: build-image-amd64
build-image-amd64: ARCHITECTURE=amd64
build-image-amd64: build-all

PHONY: build-image-arm64
build-image-arm64: ARCHITECTURE=arm64
build-image-arm64: build-all

# Tagging targets
.PHONY: tag-all tag-micro tag-minimal tag-full tag
tag-all: tag-micro tag-minimal tag-full

tag-micro:
	$(eval BUILD_ID := $(call get_build_id,$(IMAGE_NAME)-micro,$(ARCHITECTURE)))
	$(call tag_target,$(IMAGE_NAME)-micro,$(BUILD_ID))

tag-minimal:
	$(eval BUILD_ID := $(call get_build_id,$(IMAGE_NAME)-minimal,$(ARCHITECTURE)))
	$(call tag_target,$(IMAGE_NAME)-minimal,$(BUILD_ID))

tag-full:
	$(eval BUILD_ID := $(call get_build_id,$(IMAGE_NAME),$(ARCHITECTURE)))
	$(call tag_target,$(IMAGE_NAME),$(BUILD_ID))

tag: tag-full

.PHONY: push-all push-micro push-minimal push-full push

push-full:
	$(eval BUILD_ID := $(call get_build_id,$(IMAGE_NAME),$(ARCHITECTURE)))
	$(call push_target,$(IMAGE_NAME),$(BUILD_ID))

push: push-full

.PHONY: registry-login
registry-login:
	@test "${REGISTRY_USER}" != "" && test "${REGISTRY_TOKEN}" != "" || (echo "REGISTRY_USER and REGISTRY_TOKEN must be defined" && exit 1)
	@${CONTAINER_ENGINE} login -u="${REGISTRY_USER}" -p="${REGISTRY_TOKEN}" "$(IMAGE_REGISTRY)"

.PHONY: build-manifest
build-manifest:
	# builds the joint manifest for a new dual-arch container definition
	# we're currently just going to use the build id from the AMD version of the image to tag here
	$(eval AMD_BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME):latest-amd64))
	# $(eval ARM_BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME):latest-arm64))
	${CONTAINER_ENGINE} manifest exists $(IMAGE_URI)/$(IMAGE_NAME):$(AMD_BUILD_ID)-$(GIT_REVISION) && ${CONTAINER_ENGINE} manifest rm $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION) || true
	${CONTAINER_ENGINE} manifest create $(IMAGE_URI)/$(IMAGE_NAME):$(AMD_BUILD_ID)-$(GIT_REVISION) $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION)-amd64 # $(IMAGE_URI):$(ARM_BUILD_ID)-$(GIT_REVISION)-arm64
	${CONTAINER_ENGINE} manifest exists $(IMAGE_URI)/$(IMAGE_NAME):latest && ${CONTAINER_ENGINE} manifest rm $(IMAGE_URI):latest || true
	${CONTAINER_ENGINE} manifest create $(IMAGE_URI)/$(IMAGE_NAME):latest $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION)-amd64 # $(IMAGE_URI):$(ARM_BUILD_ID)-$(GIT_REVISION)-arm64

.PHONY: push-manifest
push-manifest:
	$(eval AMD_BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME):latest-amd64))
	${CONTAINER_ENGINE} manifest push $(IMAGE_URI)/$(IMAGE_NAME):$(AMD_BUILD_ID)-$(GIT_REVISION)
	${CONTAINER_ENGINE} manifest push $(IMAGE_URI)/$(IMAGE_NAME):latest

# Golang-related
.PHONY: go-build
go-build: mod fmt lint test build-snapshot

.PHONY: build-binary
build-binary:
	$(GOENV) go build -o build/$(PROJECT_NAME) .

.PHONY: mod
mod:
	go mod tidy

.PHONY: test
test:
	go test ./... -v $(TESTOPTS)

# TODO: Set this up
.PHONY: coverage
coverage:
	hack/codecov.sh

.PHONY: lint
lint: 
	$(GOPATH)/bin/golangci-lint run --timeout 5m

.PHONY: release
release: 
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is undefined)
endif
	goreleaser check --config $(GORELEASER_CONFIG)
	goreleaser release --clean --config $(GORELEASER_CONFIG) --parallelism $(GORELEASER_CORES) $(GORELEASER_ADDITIONAL_ARGS)

.PHONY: build-snapshot
build-snapshot:
	goreleaser build --clean --snapshot --single-target=true --config $(GORELEASER_CONFIG)

.PHONY: fmt
fmt:
	gofmt -s -l -w cmd pkg utils

.PHONY: clean
clean:
	rm -rf \
		build/*
		dist/*
