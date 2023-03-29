#!/usr/bin/env make

PROJECT = moby-compose

ifeq ($(CROSS),true)
CC := $(shell hack/cross/c/cc.sh)
export CC
endif

COMMIT = $(shell gunzip -c $(SOURCE) | git get-tar-commit-id || >&2 echo "unknown commit id")
export COMMIT

VERSION ?= $(shell cat VERSION)
export VERSION

REVISION ?= $(shell cat REVISION)
export REVISION

BASE_PKG_VERSION := $(shell echo $(VERSION) | tr - '~')
DISTRO_VERSION := $(shell source /etc/os-release; echo $${ID}$${VERSION_ID})
PKG_VERSION := $(BASE_PKG_VERSION)-$(DISTRO_VERSION)u$(REVISION)

SOURCE ?= $(PROJECT)-$(VERSION).tar.gz
WORKDIR := $(CURDIR)

src: $(SOURCE)

$(SOURCE):
	mkdir -p src/$(PROJECT) && cd src/$(PROJECT); \
	git init . || true; \
	git remote rm origin || true; \
	git remote add origin $(shell cat REPO) \
	&& git fetch origin $(COMMIT)
	git checkout $(COMMIT) && \
	git archive HEAD | gzip -9 > $(WORKDIR)/$@


.PHONY: env
env:
	@echo PKG_VERSION: 	$(PKG_VERSION)
	@echo COMMIT: 		$(COMMIT)
	@echo VERSION: 		$(VERSION)
	@echo REVISION: 	$(REVISION)
