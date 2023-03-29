#!/usr/bin/env make

PROJECT = moby-cli

ifeq ($(CROSS),true)
CC := $(shell hack/cross/c/cc.sh)
export CC
endif

COMMIT = $(shell gunzip -c $(SOURCE) | git get-tar-commit-id || >&2 echo "unknown commit id")
export COMMIT
export GITCOMMIT = $(COMMIT)

VERSION ?= $(shell cat VERSION)
export VERSION

REVISION ?= $(shell cat REVISION)
export REVISION

BUILDTIME = $(shell TZ=UTC date -u --date="@$(SOURCE_DATE_EPOCH)")
ifeq ($(BUILDTIME),)
BUILDTIME = unknown
endif
export BUILDTIME
export SOURCE_DATE_EPOCH



BASE_PKG_VERSION := $(shell echo $(VERSION) | tr - '~')
DISTRO_VERSION := $(shell source /etc/os-release; echo $${ID}$${VERSION_ID})
PKG_VERSION := $(BASE_PKG_VERSION)-$(DISTRO_VERSION)u$(REVISION)

SOURCE ?= $(PROJECT)-$(VERSION).tar.gz

.PHONY: env
env:
	@echo PKG_VERSION: 	$(PKG_VERSION)
	@echo COMMIT: 		$(COMMIT)
	@echo VERSION: 		$(VERSION)
	@echo REVISION: 	$(REVISION)
	@echo BUILD_TIME: 	$(BUILDTIME)

