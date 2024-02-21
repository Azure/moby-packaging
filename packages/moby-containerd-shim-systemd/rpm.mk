#!/usr/bin/make -f
.PHONY: rpm
export GOGC=off

rpm:
	if [ -n "$$GCC_VERSION" ] && [ -n "$$GCC_ENV_VILE" ]; then\
		source "$$GCC_ENV_VILE"; \
		fi; \
		cd src && make build

