deb: runc man/man8

runc:
	cd src && \
	$(MAKE) runc BUILDTAGS='seccomp' VERSION="${VERSION}-${REVISION}"

man/man8:
	cd src && \
	$(MAKE) man && $(MAKE) install-man DESTDIR= MANDIR=$(CURDIR)/man
