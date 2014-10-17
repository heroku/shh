#!/usr/bin/env make -f

VERSION := $(shell cat config.go  | grep VERSION | cut -d \" -f 2)

tempdir        := $(shell mktemp -d)
controldir     := $(tempdir)/DEBIAN
installpath    := $(tempdir)/usr/bin
buildpath      := .build
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

deb: bin/shh bin/shh-value
	echo "making deb"
	mkdir -p -m 0755 $(controldir)
	echo "$$DEB_CONTROL" > $(controldir)/control
	mkdir -p $(installpath)
	install bin/shh $(installpath)/shh
	install bin/shh-value $(installpath)/shh-value
	fakeroot dpkg-deb --build $(tempdir) .
	rm -rf $(tempdir)

clean:
	rm -rf $(buildpath)
	rm -f shh*.deb

bin/shh:
	git clone git://github.com/kr/heroku-buildpack-go.git $(buildpath)
	$(buildpath)/bin/compile . $(buildpackcache)

bin/shh-value: 
	git clone git://github.com/kr/heroku-buildpack-go.git $(buildpath)
	$(buildpath)/bin/compile . $(buildpackcache)
