#!/usr/bin/make -f
.PHONY: rpm
export GOGC=off

rpm:
	cd src && make build
