#!/usr/bin/make -f

deb:
	cd src \
		&& go build -mod=vendor \
			-ldflags "-X github.com/docker/buildx/version.Version=${VERSION}-${REVISION} -X github.com/docker/buildx/version.Revision=${COMMIT} -X github.com/docker/buildx/version.Package=github.com/docker/buildx" \
			-o docker-buildx \
			./cmd/buildx
