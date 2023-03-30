.PHONY: mariner2-rpm rpm deb rpm/% deb/%

mariner2-rpm: rpm

rpm deb:
	$(MAKE) -f $(@).mk $@

rpm/%:
	$(MAKE) -f rpm.mk $*

deb/%:
	$(MAKE) -f deb.mk $*

