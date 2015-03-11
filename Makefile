GO_LINKER_SYMBOL := "github.com/heroku/shh.version"

all: test

test:
	go test ./...

install: glv
	go install -a -ldflags "-X ${GO_LINKER_SYMBOL} ${GO_LINKER_VALUE}" ./...

update-deps: godep
	godep save -r ./...

godep:
	go get -u github.com/tools/godep

gox:
	go get -u github.com/mitchellh/gox

debs: gox glv
	$(eval TMP := $(shell mktemp -d -t shh.XXXXX))
	$(eval LINUX_AMD64 := ${TMP}/linux/amd64)
	$(eval DEB_ROOT := ${LINUX_AMD64}/DEBIAN)
	$(eval VERSION := $(shell echo ${GO_LINKER_VALUE} | sed s/^v//))
	gox -osarch="linux/amd64" -output="${TMP}/{{.OS}}/{{.Arch}}/usr/bin/{{.Dir}}" -ldflags "-X ${GO_LINKER_SYMBOL} ${GO_LINKER_VALUE}" ./...
	mkdir -p ${DEB_ROOT}
	cat misc/DEBIAN.control | sed s/{{VERSION}}/${VERSION}/ > ${DEB_ROOT}/control
	dpkg-deb -Zgzip -b ${LINUX_AMD64} shh_${VERSION}_amd64.deb
	rm -rf ${TMP}

glv:
	$(eval GO_LINKER_VALUE := $(shell git describe --tags --always))

