#!/usr/bin/env make -f

VERSION := 0.0.7

tempdir        := $(shell mktemp -d)
controldir     := $(tempdir)/DEBIAN
installpath    := $(tempdir)/usr/bin
buildpath      := .build
buildpackpath  := $(buildpath)/pack
buildpackcache := $(buildpath)/cache

define DEB_CONTROL
Package: shh
Version: $(VERSION)
Architecture: amd64
Maintainer: "Edward Muller" <edward@heroku.com>
Section: heroku
Priority: optional
Description: Systems statistics to formatted log lines
endef
export DEB_CONTROL

deb: build
	mkdir -p -m 0755 $(controldir)
	echo "$$DEB_CONTROL" > $(controldir)/control
	mkdir -p $(installpath)
	install bin/shh $(installpath)/shh
	fakeroot dpkg-deb --build $(tempdir) .
	rm -rf $(tempdir)

clean:
	rm -rf $(buildpath)
	rm -f shh*.deb

build: $(buildpackpath)/bin
	$(buildpackpath)/bin/compile . $(buildpackcache)

$(buildpackcache):
	mkdir -p $(buildpath)
	mkdir -p $(buildpackcache)
	curl -O http://codon-buildpacks.s3.amazonaws.com/buildpacks/fabiokung/go-git-only.tgz
	mv go-git-only.tgz $(buildpath)

$(buildpackpath)/bin: $(buildpackcache)
	mkdir -p $(buildpackpath)
	tar -C $(buildpackpath) -zxf $(buildpath)/go-git-only.tgz
