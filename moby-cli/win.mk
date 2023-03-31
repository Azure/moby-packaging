#!/usr/bin/make -f

BUILDTIME = $(shell TZ=UTC date -u --date="@$(SOURCE_DATE_EPOCH)")
export BUILDTIME
export CROSS := true

export GOPATH := $(CURDIR)/.gopath
export GOCACHE := $(GOPATH)/.cache
export GOPROXY := off
export GO111MODULE := off
# export GOFLAGS := -trimpath
export GOGC := off
export CGO_ENABLED ?= $(shell if [ "$$TARGETARCH" = "amd64" ]; then echo 1; else echo 0; fi)

$(GOPATH)/src/github.com/docker/cli:
	mkdir -p $(@D)
	ln -s $(CURDIR)/src $(@)

win: $(GOPATH)/src/github.com/docker/cli
	cd src && \
		QUILT_PATCHES="$(CWD)/patches" quilt push -a; \
		if [ -z "$$WINDRES" ] && [ "$$CGO_ENABLED" = "1" ]; then \
			export WINDRES="x86_64-w64-mingw32-windres"; \
		fi; \
		DISABLE_WARN_OUTSIDE_CONTAINER=1 make binary && \
		mv build/docker-windows*.exe build/docker.exe
