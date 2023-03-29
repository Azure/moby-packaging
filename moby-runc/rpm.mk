rpm: runc man/man8
.PHONY: rpm

runc:
	cd src && \
	$(MAKE) runc BUILDTAGS='seccomp'

man/man8:
	cd src && \
	$(MAKE) man && $(MAKE) install-man DESTDIR= MANDIR=$(CURDIR)/man
