CONTAINER_ENGINE:=$(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null)
ifeq ($(CONTAINER_ENGINE),)
 $(error ERROR: container engine not foumd, install podman or docker and run make again)
endif

REGISTRY_USER?=$(QUAY_USER)
REGISTRY_TOKEN?=$(QUAY_TOKEN)

IMAGE_REGISTRY?=quay.io
IMAGE_REPOSITORY?=app-sre
IMAGE_NAME?=ocm-container
IMAGE_URI=$(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)/$(IMAGE_NAME)
GIT_REVISION=$(shell git rev-parse --short=7 HEAD)
TAG?=latest
BUILD_ARGS?=
ARCHITECTURE?=$(shell arch)

.Phony: checkEnv
checkEnv:
	@test "${CONTAINER_ENGINE}" != "" || (echo "CONTAINER_ENGINE must be defined" && exit 1)
	@${CONTAINER_ENGINE} version || (echo "CONTAINER_ENGINE must be installed and in PATH" && exit 1)

.PHONY: init
init:
	bash init.sh

.PHONY: build
build:
	@${CONTAINER_ENGINE} build $(BUILD_ARGS) -t $(IMAGE_NAME):$(TAG) .

.PHONY: build-image-amd64
build-image-amd64:
	@${CONTAINER_ENGINE} build $(BUILD_ARGS) --platform=linux/amd64 -t $(IMAGE_NAME):$(TAG)-amd64 .

.PHONY: build-image-arm64
build-image-arm64:
	@${CONTAINER_ENGINE} build $(BUILD_ARGS) --platform=linux/arm64 -t $(IMAGE_NAME):$(TAG)-arm64 .

.PHONY: registry-login
registry-login:
	@test "${REGISTRY_USER}" != "" && test "${REGISTRY_TOKEN}" != "" || (echo "REGISTRY_USER and REGISTRY_TOKEN must be defined" && exit 1)
	@${CONTAINER_ENGINE} login -u="${REGISTRY_USER}" -p="${REGISTRY_TOKEN}" "$(IMAGE_REGISTRY)"

.PHONY: tag
tag:
	$(eval BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME):$(TAG)))
	# Our image tag uses the format sha256: starting our slice later to exclude that
	${CONTAINER_ENGINE} tag $(IMAGE_NAME):$(TAG) $(IMAGE_URI):$(BUILD_ID)-$(GIT_REVISION)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} tag $(IMAGE_NAME):$(TAG) $(IMAGE_URI):$(BUILD_ID)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} tag $(IMAGE_NAME):$(TAG) $(IMAGE_URI):latest-$(ARCHITECTURE)

.PHONY: push
push:
	$(eval BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME):$(TAG)))
	${CONTAINER_ENGINE} push $(IMAGE_URI):$(BUILD_ID)-$(GIT_REVISION)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} push $(IMAGE_URI):$(BUILD_ID)-$(ARCHITECTURE)
	${CONTAINER_ENGINE} push $(IMAGE_URI):latest-$(ARCHITECTURE)

.PHONY: build-manifest
build-manifest:
	# builds the joint manifest for a new dual-arch container definition
	# we're currently just going to use the build id from the AMD version of the image to tag here
	$(eval AMD_BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME):latest-amd64))
	# $(eval ARM_BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME):latest-arm64))
	${CONTAINER_ENGINE} manifest exists $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION) && ${CONTAINER_ENGINE} manifest rm $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION) || true
	${CONTAINER_ENGINE} manifest create $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION) $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION)-amd64 # $(IMAGE_URI):$(ARM_BUILD_ID)-$(GIT_REVISION)-arm64
	${CONTAINER_ENGINE} manifest exists $(IMAGE_URI):latest && ${CONTAINER_ENGINE} manifest rm $(IMAGE_URI):latest || true
	${CONTAINER_ENGINE} manifest create $(IMAGE_URI):latest $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION)-amd64 # $(IMAGE_URI):$(ARM_BUILD_ID)-$(GIT_REVISION)-arm64

.PHONY: push-manifest
push-manifest:
	$(eval AMD_BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME):latest-amd64))
	${CONTAINER_ENGINE} manifest push $(IMAGE_URI):$(AMD_BUILD_ID)-$(GIT_REVISION)
	${CONTAINER_ENGINE} manifest push $(IMAGE_URI):latest

.PHONY: tag-n-push
tag-n-push: registry-login tag push


.Phony: test
test: go test ./...
