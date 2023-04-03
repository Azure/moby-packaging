.PHONY: rpm deb rpm/% deb/%

rpm deb:
	$(MAKE) -f $(@).mk $@

rpm/%:
	$(MAKE) -f rpm.mk $*

deb/%:
	$(MAKE) -f deb.mk $*

