.PHONY: mariner2-rpm rpm deb rpm/% deb/% win win/%

mariner2-rpm: rpm

rpm deb win:
	$(MAKE) -f $(@).mk $@

rpm/%:
	$(MAKE) -f rpm.mk $*

deb/%:
	$(MAKE) -f deb.mk $*

win/%:
	$(MAKE) -f win.mk $*
