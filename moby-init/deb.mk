.PHONY: deb
deb: tini-static

tini-static:
	mkdir -vp src/build
	cd src/build && \
		cmake .. && \
		make tini-static
