export DOCKER_BUILDKIT := 1
export DOCKER_BUILDKIT

SHELL := /usr/bin/env bash

DOCKER_BUILDX ?= docker buildx

ifeq ($(V), 1)
PROGRESS := --progress=plain
endif

export V

OUTPUT ?= $(CURDIR)/bundles
export OUTPUT

COMPONENT_TARGETS = $(wildcard moby-*)

all:
	docker buildx bake all

.PHONY: clean
clean:
	if [ -n "$(DISTRO)" ]; then \
		$(MAKE) -C tests clean DISTRO=$(DISTRO); \
		rm -rf $(OUTPUT)/$(DISTRO); \
	else \
		$(MAKE) -C tests clean; \
		rm -rf {bin,bundles}; \
	fi


.PHONY: $(COMPONENT_TARGETS)
ifdef CROSS
export CROSS
endif

ifdef TARGETARCH
export TARGETARCH
endif

ifdef GO_IMAGE
export GO_IMAGE
endif

ifdef BUILDKIT_MULTI_PLATFORM
export BUILDKIT_MULTI_PLATFORM
endif

ifdef TARGETARCH
export TARGETARCH
endif


PULL ?= true

$(COMPONENT_TARGETS):
	if [ "$(UPDATE)" = "1" ]; then (cd $@ && ./update.sh); fi
	$(DOCKER_BUILDX) bake $(PROGRESS) $@

_LOCAL_PKG_DIR := $(OUTPUT)/$(DISTRO)
ifneq ($(wildcard $(_LOCAL_PKG_DIR)/linux_*),)
	_LOCAL_PKG_DIR := $(wildcard $(_LOCAL_PKG_DIR)/linux_*)
endif
ifneq ($(wildcard $(_LOCAL_PKG_DIR)/*),)
export LOCAL_PKG_DIR ?= $(_LOCAL_PKG_DIR)
endif

.PHONY: test
test:
	$(MAKE) -s -C tests test DISTRO=$(DISTRO) LOCAL_PKG_DIR=$(LOCAL_PKG_DIR)

.PHONY: test/%
test/%:
	$(MAKE) -s -C tests $* DISTRO=$(DISTRO) LOCAL_PKG_DIR=$(LOCAL_PKG_DIR)

update:
	@ ./update.sh
 