#!/usr/bin/make -f

include base.mk

export GOPATH := $(CURDIR)/.gopath
export GOCACHE := $(GOPATH)/.cache
export GOPROXY := off
export GO111MODULE := off
# export GOFLAGS := -trimpath
export GOGC := off
export CGO_ENABLED := 1

COMMIT = $(shell gunzip -c containerd.tar.gz | git get-tar-commit-id || >&2 echo "unknown commit id")
export COMMIT
export GITCOMMIT = $(COMMIT)

containerd_src_dir := $(GOPATH)/src/github.com/containerd/containerd
hcs_src_dir := $(GOPATH)/src/github.com/microsoft/hcsshim

containerd.tar.gz: $(SOURCE)
	tar -xvf $(SOURCE)

hcs.tar.gz: $(SOURCE)
	tar -xvf $(SOURCE)

$(containerd_src_dir): containerd.tar.gz
	mkdir -p $@
	tar -C $@ -xf $<

$(hcs_src_dir): hcs.tar.gz
	mkdir -p $@
	tar -C $@ -xf $<


.PHONY: binaries containerd_binaries hcs_binaries zip

OUTPUT := $(PROJECT)-$(PKG_VERSION).$(TARGETARCH)$(TARGETVARIANT).zip

BUILD_DIR := $(CURDIR)/bundles


containerd_binaries: $(containerd_src_dir)
	cd $(containerd_src_dir) && \
	$(MAKE) VERSION=$(VERSION) REVISION=$(COMMIT) binaries && \
	rm bin/containerd-stress.exe && \
	mkdir -p $(BUILD_DIR) && \
	mv bin/* $(BUILD_DIR)/

hcs_binaries: $(hcs_src_dir)
	cd $(hcs_src_dir) && \
	mkdir -p $(BUILD_DIR) && \
	GO111MODULE=on go build -mod=vendor -o "$(BUILD_DIR)/containerd-shim-runhcs-v1.exe" ./cmd/containerd-shim-runhcs-v1

binaries: containerd_binaries hcs_binaries

zip: $(OUTPUT)

$(OUTPUT): binaries
	zip -r -j "$@" $(BUILD_DIR)
