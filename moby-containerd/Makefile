.PHONY: deb win rpm _rpm _hcsbinaries _binaries _man

export GOPROXY := direct
export GO111MODULE := on
export GOFLAGS := -trimpath
export GOGC := off

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

_hcsbinaries:
	cd src/hcs-shim && \
		mkdir -p $(CURDIR)/src/bin && \
		GO111MODULE=on go build -mod=vendor -o "bin/containerd-shim-runhcs-v1.exe" ./cmd/containerd-shim-runhcs-v1
