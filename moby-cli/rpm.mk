rpm: _binary _man

BUILDTIME = $(shell TZ=UTC date -u --date="@$(SOURCE_DATE_EPOCH)")
export BUILDTIME
.PHONY: rpm

$(GOPATH)/src/github.com/docker/cli:
	mkdir -p $(@D)
	ln -s $(CURDIR)/src $(@)

_binary: $(GOPATH)/src/github.com/docker/cli
	cd src && DISABLE_WARN_OUTSIDE_CONTAINER=1 \
		make \
			LDFLAGS='' \
			VERSION='$(VERSION)-$(REVISION)' \
			GITCOMMIT='$(COMMIT)' \
			BUILDTIME='$(BUILDTIME)' \
			dynbinary

# https://github.com/docker/cli/blob/v19.03.5/scripts/docs/generate-man.sh
# (replacing hard-coded "/tmp/gen-manpages" with "debian/tmp/gen-manpages")
_man:
	export PATH='$(GOPATH)/bin':"$$PATH"; \
	export SKIP_GOLANG_WRAPPER=1; \
	export CC=""; \
	cd src && make manpages
