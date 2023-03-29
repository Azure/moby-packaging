#!/usr/bin/make -f
.PHONY: rpm

rpm:
	cd src \
		&& make VERSION=$(VERSION)-$(REVISION) DESTDIR=bin/ build
