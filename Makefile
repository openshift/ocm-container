PROJECT_NAME := ocm-container

# Go binary detection
GO_BIN := $(shell command -v go 2>/dev/null)
ifeq ($(GO_BIN),)
$(error ERROR: go binary not found. Ensure go is installed and run make again)
endif

# Container engine detection
CONTAINER_ENGINE := $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null)
ifeq ($(CONTAINER_ENGINE),)
$(error ERROR: container engine not found. Ensure podman or docker are installed and run make again)
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

# Default target is to build the full container image and tag it
# as `ocm-container:latest` for local use.  The default target is 
# intended for human use, outside of the CI/CD pipeline.
default: build tag

.Phony: check
check:
	@echo "Checking environment configuration..."
	@$(MAKE) check-env
	@echo "Checking GitHub API quota..."
	@$(MAKE) check-github-quota

# Helper to display environment configuration
.PHONY: check-env
check-env:
	@echo "==================================="
	@echo " OCM Container Makefile Environment"
	@echo "==================================="
	@echo
	@echo "Project Configuration:"
	@printf "  %-20s %s\n" "Project Name:" "$(PROJECT_NAME)"
	@printf "  %-20s %s\n" "Container Engine:" "$(CONTAINER_ENGINE)"
	@printf "  %-20s %s\n" "Git Revision:" "$(GIT_REVISION)"
	@echo
	@echo "Registry & Image Settings:"
	@printf "  %-20s %s\n" "Registry:" "$(IMAGE_REGISTRY)"
	@printf "  %-20s %s\n" "Repository:" "$(IMAGE_REPOSITORY)"
	@printf "  %-20s %s\n" "Image Name:" "$(IMAGE_NAME)"
	@printf "  %-20s %s\n" "Image URI:" "$(IMAGE_URI)"
	@printf "  %-20s %s\n" "Tag:" "$(TAG)"
	@printf "  %-20s %s\n" "Registry User:" "$(REGISTRY_USER)"
	@printf "  %-20s %s\n" "Registry User:" "UNSET"
ifdef REGISTRY_TOKEN
	@printf "  %-20s %s\n" "Registry Token:" "SET (hidden)"
else
	@printf "  %-20s %s\n" "Registry Token:" "UNSET"
endif
	@echo
	@echo "Build Configuration:"
	@printf "  %-20s %s\n" "Architecture:" "$(ARCHITECTURE) (raw: $(RAW_ARCHITECTURE))"
	@printf "  %-20s %s\n" "Build Args:" "$(BUILD_ARGS)"
	@printf "  %-20s %s\n" "Cache:" "$(CACHE)"
ifdef GITHUB_TOKEN
	@printf "  %-20s %s\n" "GitHub Token:" "SET (hidden)"
ifdef GITHUB_BUILD_ARGS
	@printf "  %-20s %s\n" "GitHub Build Args:" "SET (hidden)"
else
	@printf "  %-20s %s\n" "GitHub Build Args:" "UNSET"
endif
else
	@printf "  %-20s %s\n" "GitHub Token:" "UNSET"
	@printf "  %-20s %s\n" "GitHub Build Args:" "UNSET"
endif
	@echo
	@echo "Go Environment:"
	@printf "  %-20s %s\n" "Go Binary:" "$(GO_BIN)"
	@printf "  %-20s %s\n" "GOOS:" "$(GOOS)"
	@printf "  %-20s %s\n" "GOARCH:" "$(GOARCH)"
	@printf "  %-20s %s\n" "GOPATH:" "$(GOPATH)"
	@printf "  %-20s %s\n" "GO111MODULE:" "$(GO111MODULE)"
	@printf "  %-20s %s\n" "GOPROXY:" "$(GOPROXY)"
	@printf "  %-20s %s\n" "CGO_ENABLED:" "$(CGO_ENABLED)"
	@printf "  %-20s %s\n" "Test Options:" "$(TESTOPTS)"
	@echo
	@echo "Tool Versions:"
	@printf "  %-20s %s\n" "golangci-lint:" "$(GOLANGCI_LINT_VERSION)"
	@printf "  %-20s %s\n" "goreleaser:" "$(GORELEASER_VERSION)"
	@printf "  %-20s %s\n" "goreleaser config:" "$(GORELEASER_CONFIG)"
	@printf "  %-20s %s\n" "goreleaser cores:" "$(GORELEASER_CORES)"
	@printf "  %-20s %s\n" "goreleaser args:" "$(GORELEASER_ADDITIONAL_ARGS)"

# Helper to check GitHub quota for GITHUB_TOKEN
.PHONY: check-github-quota
check-github-quota:
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is not set)
endif

	@echo "Checking GitHub API quota..."
	@curl -s -H "Authorization: token $(GITHUB_TOKEN)" https://api.github.com/rate_limit | jq '.rate'


# Helper macro: $(call push_manifest,<image name>)
# Pushes the manifest for the specified image name
define push_manifest
	@echo "Pushing manifest for image: $(IMAGE_URI)/$(1):latest"
	@if ! ${CONTAINER_ENGINE} manifest exists $(IMAGE_URI)/$(1):latest; then \
		echo "ERROR: Manifest for $(IMAGE_URI)/$(1):latest does not exist"; \
		exit 1; \
	fi
	${CONTAINER_ENGINE} manifest push $(IMAGE_URI)/$(1):latest
endef

# Helper macro: $(call remove_manifest,<image name>
# Removes the manifest for the specified image name
# The `|| true` ensures that if the manifest does not exist, the command does not fail
define remove_manifest
	@echo "Removing manifest for image: $(IMAGE_URI)/$(1):latest"
	${CONTAINER_ENGINE} manifest exists $(IMAGE_URI)/$(1):latest && ${CONTAINER_ENGINE} manifest rm $(IMAGE_URI)/$(1):latest || true
endef

# Helper macro: $(call build_target,<image name>,<architecture>)
# Builds the container image for the specified target and architecture
define build_target
	@echo "Building image: $(1) for architecture: $(2) with manifest $(IMAGE_URI)/$(1):latest"
	$(eval BUILD_FLAGS := --target=$(1) --platform=$(2) $(CACHE) $(BUILD_ARGS))
	$(eval BUILD_FLAGS += $(if $(GITHUB_BUILD_ARGS),$(GITHUB_BUILD_ARGS)))
	$(eval BUILD_FLAGS += -f Containerfile)
	$(CONTAINER_ENGINE) build --jobs=2 --manifest=$(IMAGE_URI)/$(1):latest $(BUILD_FLAGS) -t $(1):$(2) .
endef

# Helper macro: $(call build_local_target,<image name>,<architecture>)
# Builds the container image for local use without manifest
define build_local_target
	@echo "Building local image: $(1) for architecture: $(2) (without manifest)"
	$(eval BUILD_FLAGS := --target=$(1) --platform=$(2) $(CACHE) $(BUILD_ARGS))
	$(eval BUILD_FLAGS += $(if $(GITHUB_BUILD_ARGS),$(GITHUB_BUILD_ARGS)))
	$(eval BUILD_FLAGS += -f Containerfile)
	$(CONTAINER_ENGINE) build $(BUILD_FLAGS) -t $(1):latest .
endef

# Helper macro: $(call tag_target,<image name>, <build id>)
define tag_target
	${CONTAINER_ENGINE} tag $(1):$(ARCHITECTURE) $(IMAGE_URI)/$(1):$(2)-$(GIT_REVISION)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} tag $(1):$(ARCHITECTURE) $(IMAGE_URI)/$(1):$(2)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} tag $(1):$(ARCHITECTURE) $(IMAGE_URI)/$(1):latest-$(ARCHITECTURE)
endef

# Helper macro: $(call tag_local_target,<image name>,<build id>)
define tag_local_target
	${CONTAINER_ENGINE} tag $(1):$(ARCHITECTURE) $(1):latest
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

build-micro: check
	@$(call build_target,$(IMAGE_NAME)-micro,$(ARCHITECTURE))

build-minimal: check
	@$(call build_target,$(IMAGE_NAME)-minimal,$(ARCHITECTURE))

build-full: check
	@$(call build_target,$(IMAGE_NAME),$(ARCHITECTURE))

build-full-local: check
	@$(call build_local_target,$(IMAGE_NAME),$(ARCHITECTURE))

# The default build target is for human use, outside of the CI/CD pipeline
build: build-full-local

.PHONY: build-image-amd64
build-image-amd64: ARCHITECTURE=amd64
build-image-amd64: build-all

.PHONY: build-image-arm64
build-image-arm64: ARCHITECTURE=arm64
build-image-arm64: build-all

# Tagging targets
.PHONY: tag-all tag-micro tag-minimal tag-full tag-full-local tag
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

# "tag-full-local" is the default full target,  to ensure "ocm-container:latest" exists on the local system
# Intended for humans running manually, outside of the CI/CD pipeline
# This is called when running the default `make` or `make build` commands
tag-full-local:
	$(eval BUILD_ID := $(call get_build_id,$(IMAGE_NAME),$(ARCHITECTURE)))
	$(call tag_local_target,$(IMAGE_NAME),$(BUILD_ID))

tag: tag-full-local

# Push targets
.PHONY: push-all push-micro push-minimal push-full push
push-all: push-micro push-minimal push-full

push-micro:
	$(eval BUILD_ID := $(call get_build_id,$(IMAGE_NAME)-micro,$(ARCHITECTURE)))
	$(call push_target,$(IMAGE_NAME)-micro,$(BUILD_ID))

push-minimal:
	$(eval BUILD_ID := $(call get_build_id,$(IMAGE_NAME)-minimal,$(ARCHITECTURE)))
	$(call push_target,$(IMAGE_NAME)-minimal,$(BUILD_ID))

push-full:
	$(eval BUILD_ID := $(call get_build_id,$(IMAGE_NAME),$(ARCHITECTURE)))
	$(call push_target,$(IMAGE_NAME),$(BUILD_ID))

push: push-full

.PHONY: registry-login
registry-login:
	@test "${REGISTRY_USER}" != "" && test "${REGISTRY_TOKEN}" != "" || (echo "REGISTRY_USER and REGISTRY_TOKEN must be defined" && exit 1)
	@${CONTAINER_ENGINE} login -u="${REGISTRY_USER}" -p="${REGISTRY_TOKEN}" "$(IMAGE_REGISTRY)"

# Removes any existing manifest for the three images
# This is used to clean up before building a new joint manifest
.PHONY: remove-manifests
remove-manifests:
	$(call remove_manifest,$(IMAGE_NAME)-micro)
	$(call remove_manifest,$(IMAGE_NAME)-minimal)
	$(call remove_manifest,$(IMAGE_NAME))

.PHONY: push-manifests push-manifest-all push-manifest-micro push-manifest-minimal push-manifest-full
push-manifest-all:
	$(call push_manifest,$(IMAGE_NAME)-micro)
	$(call push_manifest,$(IMAGE_NAME)-minimal)
	$(call push_manifest,$(IMAGE_NAME))

push-manifest-micro:
	$(call push_manifest,$(IMAGE_NAME)-micro)

push-manifest-minimal:
	$(call push_manifest,$(IMAGE_NAME)-minimal)

push-manifest-full:
	$(call push_manifest,$(IMAGE_NAME))

push-manifests: push-manifest-all

# CI helper targets
.PHONY: pr-check check-image-build release-image
# TODO: Add golang build/tests here (onboard project to boilerplate?)
pr-check: check-image-build

check-image-build:
	@echo "Checking image build..."
	@bash .ci/pull-request-check.sh

release-image:
	@echo "Running release image build..."
	@bash .ci/release-build.sh

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

.PHONY: release-binary
release-binary:
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
