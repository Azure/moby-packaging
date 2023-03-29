#!/usr/bin/make -f

include base.mk

export GOPATH := $(CURDIR)/.gopath
export GOCACHE := $(GOPATH)/.cache
export GOPROXY := off
export GO111MODULE := off
# export GOFLAGS := -trimpath
export GOGC := off
export CGO_ENABLED := 1
SHELL := /bin/bash

.PHONY: win bundles/binary-daemon/dockerd.exe

win: bundles/binary-daemon/dockerd.exe

$(GOPATH)/src/github.com/docker/docker:
	mkdir -p $(@D)
	ln -s $(CURDIR)/src $(@)

bundles/binary-daemon/dockerd.exe: $(GOPATH)/src/github.com/docker/docker
	set -e; \
	cd src && \
	QUILT_PATCHES="$(PWD)/patches" quilt push -a || let ec=$$?; \
	[ ! $$ec -eq 0 ] && [ ! $$ec -eq 2 ] && exit $$ec; \
	PKG_CONFIG_PATH="$(pkg-config --variable pc_path pkg-config)"; \
	for i in $(find /usr/lib -name 'pkgconfig'); do PKG_CONFIG_PATH="${PKG_CONFIG_PATH}:$i"; done; \
	export PKG_CONFIG_PATH; \
	GOOS=windows hack/make.sh binary
