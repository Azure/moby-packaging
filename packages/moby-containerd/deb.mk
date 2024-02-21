#!/usr/bin/make -f

# circumvent a few problematic (for Debian) Go features inspired by dh-golang
export GOPROXY := direct
export GO111MODULE := on
export GOFLAGS := -trimpath
export GOGC := off
.PHONY: deb _binaries _man

SHELL := $(shell which bash)

deb: _binaries _man

_binaries:
	cd src \
		&& $(MAKE) binaries \
			VERSION='$(VERSION)-$(REVISION)' \
			REVISION='$(COMMIT)'

_man:
	mkdir -vp /man
	export SKIP_GOLANG_WRAPPER=1; \
	export CC=""; \
	cd src \
		&& [ -n "$$(ls man/*)" ] || $(MAKE) man && $(MAKE) install-man DESTDIR= MANDIR=$(CURDIR)/man
