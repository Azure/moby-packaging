# circumvent a few problematic (for Debian) Go features inspired by dh-golang
.PHONY: rpm
export GOPROXY := off
export GO111MODULE := off
# export GOFLAGS := -trimpath
export GOGC := off
export CGO_ENABLED := 1
export DOCKER_BUILDTAGS := exclude_graphdriver_btrfs

rpm: $(GOPATH)/src/github.com/docker/docker bundles/dynbinary-daemon/dockerd libnetwork/docker-proxy

$(GOPATH)/src/github.com/docker/docker:
	mkdir -p $(@D)
	ln -s $(CURDIR)/src $(@)

libnetwork/docker-proxy: # (from libnetwork)
	cd $(GOPATH)/src/github.com/docker/docker && CGO_ENABLED=false CC="" go build \
		-o libnetwork/docker-proxy \
		./cmd/docker-proxy

bundles/dynbinary-daemon/dockerd: # engine
	 DOCKER_GITCOMMIT=$(COMMIT) VERSION=$(VERSION)-$(REVISION) PRODUCT=docker hack/make.sh dynbinary
