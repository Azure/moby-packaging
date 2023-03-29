.PHONY: rpm deb win rpm/% deb/% win/%

rpm deb win:
	$(MAKE) -f $(@).mk $@

rpm/%:
	$(MAKE) -f rpm.mk $*

deb/%:
	$(MAKE) -f deb.mk $*

win/%:
	$(MAKE) -f win.mk $*
