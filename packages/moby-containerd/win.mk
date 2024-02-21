#!/usr/bin/make -f

# circumvent a few problematic (for Debian) Go features inspired by dh-golang
export GOPROXY := direct
export GO111MODULE := on
export GOFLAGS := -trimpath
export GOGC := off
.PHONY: win _binaries _man

win: _binaries hcs_binaries

_binaries:
	cd src \
		&& $(MAKE) binaries \
			VERSION='$(VERSION)-$(REVISION)' \
			REVISION='$(COMMIT)'


hcs_binaries:
	cd src/hcs-shim && \
		mkdir -p $(CURDIR)/src/bin && \
		GO111MODULE=on go build -mod=vendor -o "bin/containerd-shim-runhcs-v1.exe" ./cmd/containerd-shim-runhcs-v1
