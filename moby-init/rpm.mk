.PHONY: deb
.PHONY: rpm
rpm: tini-static

tini-static:
	mkdir -vp src/build
	cd src/build && \
		cmake .. && \
		make tini-static
