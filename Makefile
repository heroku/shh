GO_LINKER_SYMBOL := "github.com/heroku/shh.version"
GO_BUILD_ENV := GOOS=linux GOARCH=amd64

all: test

test:
	go test -v ./...
	go test -v -race ./...

install: glv
	go install -v ${LDFLAGS} ./...

debs: tmp ldflags  ver
	$(eval DEB_ROOT := ${TMP}/DEBIAN)
	${GO_BUILD_ENV} go build -v -o ${TMP}/usr/bin/shh ${LDFLAGS} ./cmd/shh
	${GO_BUILD_ENV} go build -v -o ${TMP}/usr/bin/shh-value ${LDFLAGS} ./cmd/shh-value
	mkdir -p ${DEB_ROOT}
	cat misc/DEBIAN.control | sed s/{{VERSION}}/${VERSION}/ > ${DEB_ROOT}/control
	dpkg-deb -Zgzip -b ${TMP} shh_${VERSION}_amd64.deb
	rm -rf ${TMP}

glv:
	$(eval GO_LINKER_VALUE := $(shell git describe --tags --always))

ldflags: glv
	$(eval LDFLAGS := -ldflags "-X ${GO_LINKER_SYMBOL}=${GO_LINKER_VALUE}")

ver: glv
	$(eval VERSION := $(shell echo ${GO_LINKER_VALUE} | sed s/^v//))

docker: ldflags ver clean-docker-build
	${GO_BUILD_ENV} go build -v -o .docker_build/shh ${LDFLAGS} ./cmd/shh
	${GO_BUILD_ENV} go build -v -o .docker_build/shh-value ${LDFLAGS} ./cmd/shh-value
	docker build -t heroku/shh:${VERSION} ./
	${MAKE} clean-docker-build

clean-docker-build:
	rm -rf .docker_build

tmp:
	$(eval TMP := $(shell mktemp -d -t shh.XXXXX))