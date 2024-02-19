#!/usr/bin/make -f

deb:
	cd src \
		&& make VERSION=$(VERSION)-$(REVISION) DESTDIR=bin/ build
