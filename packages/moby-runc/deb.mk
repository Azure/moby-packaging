deb: runc man/man8

runc:
	cd src && \
	$(MAKE) runc BUILDTAGS='seccomp urfave_cli_no_docs' VERSION="${VERSION}-${REVISION}" COMMIT="${COMMIT}"

man/man8:
	cd src && \
	$(MAKE) man && $(MAKE) install-man DESTDIR= MANDIR=$(CURDIR)/man
