#!/usr/bin/env make

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --warn-undefined-variables
.ONESHELL:

BUILD := docker buildx build --load

ifdef LOCAL_PKG_DIR
export LOCAL_PKG_DIR
endif

ifdef INCLUDE_TESTING
export INCLUDE_TESTING
endif

ifdef TARGETARCH
platform := --platform=linux/$(TARGETARCH)
else
platform :=
endif

DOCKER_ENV := \
	-e TEST_CLI_COMMIT \
	-e TEST_CLI_VERSION \
	-e TEST_CONTAINERD_COMMIT \
	-e TEST_CONTAINERD_VERSION \
	-e TEST_ENGINE_COMMIT \
	-e TEST_ENGINE_VERSION  \
	-e TEST_RUNC_VERSION \
	-e TEST_RUNC_COMMIT \
	-e TEST_BUILDX_VERSION \
	-e TEST_BUILDX_COMMIT \
	-e TEST_COMPOSE_VERSION \
	-e TEST_RUNC_PACKAGE_VERSION \
	-e TEST_ENGINE_PACKAGE_VERSION \
	-e TEST_CONTAINERD_PACKAGE_VERSION \
	-e TEST_CLI_PACKAGE_VERSION \
	-e TEST_COMPOSE_PACKAGE_VERSION \
	-e TEST_BUILDX_PACKAGE_VERSION \
	-e TARGETARCH \

TESTDIR ?= $(CURDIR)/.test

.PHONY: img
img: $(TESTDIR)/$(DISTRO)/imageid

# Explict dependency tracking per distro
bionic jammy focal bullseye buster: $(wildcard deb/*)
rpm: rpm-fix-arch.sh
centos8 rhel8 rhel9: $(wildcard centos8/*) rpm
centos7: $(wildcard centos7/*) rpm
mariner2: $(wildcard mariner2/*) rpm

$(TESTDIR)/%/nid:
	mkdir -p $(@D)
	echo "bridge" > "$(@)"

# mariner is... weird...
# namely it has some networking issues that lowering MTU seems to improve
$(TESTDIR)/mariner2/nid:
	mkdir -p $(@D)
	docker network create --opt com.docker.network.driver.mtu=1420 test_moby_$(DISTRO)_$${RANDOM} > "$(@)"

$(TESTDIR)/mariner2/imageid: mariner2/install.sh mariner2/download-pcks.sh

$(TESTDIR)/$(DISTRO)/imageid: $(DISTRO) Dockerfile entrypoint.sh test.sh
	if [ -z "$(DISTRO)" ]; then \
		>&2 echo Must set DISTRO; \
		exit 1; \
	fi; \
	$(MAKE) DISTRO="$(DISTRO)" TESTDIR="$(TESTDIR)" clean/$(DISTRO) && \
	mkdir -p "$(@D)"; \
	$(BUILD) \
		--target "$(DISTRO)-test" \
		--iidfile "$(@)" \
		--build-arg INCLUDE_TESTING \
		--build-arg INCLUDE_STAGING \
		--build-arg $(DISTRO)_IMG=mcr.microsoft.com/mirror/docker/library/ \
		--label azure.com/moby/packaging/distro=$(DISTRO) \
		--label azure.com/moby/packaging/test \
		${platform} \
		.

$(TESTDIR)/$(DISTRO)/cid: $(TESTDIR)/$(DISTRO)/imageid $(TESTDIR)/$(DISTRO)/nid
	docker run -d \
		--cidfile "$(@)" \
		-t \
		--privileged \
		-e container=docker \
		--tmpfs /tmp \
		--tmpfs /run \
		--tmpfs /run/lock \
		--network "$$(cat $(TESTDIR)/$(DISTRO)/nid)" \
		-v /var/lib/docker \
		-v /var/lib/containerd \
		-v /lib/modules:/lib/modules:ro \
		${platform} \
	"$$(cat $(TESTDIR)/$(DISTRO)/imageid)"



ifneq ($(LOCAL_PKG_DIR),)
$(LOCAL_PKG_DIR): $(TESTDIR)/$(DISTRO)/cid $(wildcard $(LOCAL_PKG_DIR)/*)
	id="$$(cat $<)"; \
	docker exec "$${id}" /bin/sh -c 'rm -rf /var/pkg'; \
	docker cp "$(LOCAL_PKG_DIR)/" "$${id}:/var/pkg/"
endif


ifneq ($(LOCAL_PKG_DIR),)
$(TESTDIR)/$(DISTRO)/installed: $(TESTDIR)/$(DISTRO)/cid $(LOCAL_PKG_DIR)
else
$(TESTDIR)/$(DISTRO)/installed: $(TESTDIR)/$(DISTRO)/cid
endif
	id="$$(cat $<)"
	docker start $${id}
	docker exec \
		$(DOCKER_ENV) \
		$${id} /opt/moby/install.sh /var/pkg
	echo 1 > $@


## Tests just don't work when running under emulation (b/c qemu can't support it)
## It's still valuable to go through the install process, though.
ifeq (SKIP_TESTS, 1)
$(TESTDIR)/$(DISTRO)/TestReport-test.sh.xml: $(TESTDIR)/$(DISTRO)/installed
	@echo Skipping test suite due to SKIP_TESTS=1
else
.PHONY: $(TESTDIR)/$(DISTRO)/TestReport-test.sh.xml
$(TESTDIR)/$(DISTRO)/TestReport-test.sh.xml: $(TESTDIR)/$(DISTRO)/installed
	set +e; \
	echo "Testing $(DISTRO)"; \
	id="$$(cat $(TESTDIR)/$(DISTRO)/cid)"; \
	docker start "$${id}"; \
	docker exec \
		$(DOCKER_ENV) \
		"$${id}" bats --formatter junit -T -o /opt/moby/ /opt/moby/test.sh; \
	let ec=$$?; \
	if [ $$ec -gt 0 ]; then >&2 echo "$(DISTRO) failed"; fi; \
	docker cp "$${id}:/opt/moby/$(@F)" "$(@)"; \
	echo $@; \
	exit $$ec
endif

.PHONY: install
install: $(TESTDIR)/$(DISTRO)/installed

.PHONY: logs
logs: $(TESTDIR)/$(DISTRO)/cid
	id="$$(cat $<)";\
	docker logs "$${id}"; \
	docker exec "$${id}" journalctl -u docker; \
	docker exec "$${id}" journalctl -u containerd; \
	docker inspect "$${id}"

.PHONY: shell
shell: $(TESTDIR)/$(DISTRO)/cid
	id="$$(cat $<)"; \
	docker start "$${id}"; \
 	if [ -n "$(LOCAL_PKG_DIR)" ]; then docker cp $(LOCAL_PKG_DIR)/ "$${id}:/var/pkg/"; fi; \
 	docker exec $(DOCKER_ENV) -it "$${id}" bash

.PHONY: test
test: $(TESTDIR)/$(DISTRO)/TestReport-test.sh.xml

commit: $(TESTDIR)/$(DISTRO)/installed $(TESTDIR)/$(DISTRO)/cid
	@if [ -z "$(COMMIT_NAME)" ]; then \
		>&2 echo Must set COMMIT_NAME; \
		exit 1; \
	fi; \
	docker commit $$(cat $(TESTDIR)/$(DISTRO)/cid) $(COMMIT_NAME)

.PHONY: $(TESTDIR)/$(DISTRO)/clean
$(TESTDIR)/$(DISTRO)/clean:
	if [ -f "$(@D)/cid" ]; then \
		docker rm -fv "$$(cat $(@D)/cid)" || true; \
	fi; \
	if [ -f "$(@D)/imageid" ]; then \
		docker rmi -f "$$(cat $(@D)/imageid)"; \
	fi; \
	if [ -f "$(@D)/nid" ]; then \
		docker network rm "$$(cat $(@D)/nid)" || true; \
	fi; \
	rm -rf $(@D);

.PHONY: clean
clean:
	if [ -n "$(DISTRO)" ]; then \
		$(MAKE) "clean/$(DISTRO)"; \
		exit 0; \
	fi; \
	for i in $(TESTDIR)/*; do \
		[ ! -d "$${i}" ] && continue; \
		$(MAKE) "$${i}/clean" DISTRO="$$(basename $${i})" TESTDIR="$(TESTDIR)"; \
	done; \
	rm -rf "$(TESTDIR)"

.PHONY: clean/%
clean/%:
	$(MAKE) $(TESTDIR)/$(*)/clean DISTRO=$(*) TESTDIR=$(TESTDIR)
