.PHONY: deb win rpm _rpm _hcsbinaries _binaries _man _hcs_shim_source

export GOPROXY := direct
export GO111MODULE := on
export GOFLAGS := -trimpath
export GOGC := off

HCS_SHIM_REPO ?= https://github.com/Microsoft/hcsshim.git

deb: _binaries _man

win: _binaries _hcsbinaries

rpm: _rpm _binaries _man

_rpm: rpm
export BUILDTAGS := no_btrfs

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

_hcsbinaries: _hcs_shim_source
	cd /build/hcs-shim && \
		mkdir -p $(CURDIR)/src/bin && \
		GO111MODULE=on go build -mod=vendor -o "/build/hcs-shim/bin/containerd-shim-runhcs-v1.exe" ./cmd/containerd-shim-runhcs-v1

_hcs_shim_source:
	mkdir -p /build/hcs-shim
	cd /build/src && \
		HCS_SHIM_COMMIT="$$(awk '/Microsoft\/hcsshim/{ print $$2 }' < go.mod)" && \
		git clone --depth=1 -b "$${HCS_SHIM_COMMIT}" "$(HCS_SHIM_REPO)" /build/hcs-shim
