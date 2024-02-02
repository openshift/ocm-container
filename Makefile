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

.PHONY: init
init:
	bash init.sh

.PHONY: build
build:
	bash build.sh

.PHONY: registry-login
registry-login:
	@test "${REGISTRY_USER}" != "" && test "${REGISTRY_TOKEN}" != "" || (echo "REGISTRY_USER and REGISTRY_TOKEN must be defined" && exit 1)
	@${CONTAINER_ENGINE} login -u="${REGISTRY_USER}" -p="${REGISTRY_TOKEN}" "$(IMAGE_REGISTRY)"

.PHONY: tag
tag:
	# Our image tag uses the format sha256: starting our slice later to exclude that
	$(eval BUILD_ID=$(shell ${CONTAINER_ENGINE} image inspect --format '{{slice .ID 7 19}}' $(IMAGE_NAME)))
	${CONTAINER_ENGINE} tag $(IMAGE_NAME) $(IMAGE_URI):$(BUILD_ID)-$(GIT_REVISION)
	${CONTAINER_ENGINE} tag $(IMAGE_NAME) $(IMAGE_URI):$(BUILD_ID)
	${CONTAINER_ENGINE} tag $(IMAGE_NAME) $(IMAGE_URI):latest

.PHONY: push
push:
	${CONTAINER_ENGINE} push $(IMAGE_URI):$(BUILD_ID)-$(GIT_REVISION)
	${CONTAINER_ENGINE} push $(IMAGE_URI):$(BUILD_ID)
	${CONTAINER_ENGINE} push $(IMAGE_URI):latest

.PHONY: tag-n-push
tag-n-push: registry-login tag push
