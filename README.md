System Heuristics Herald (aka Shh)
----

Gathers and relays system metrics

## Install

    go get github.com/freeformz/shh

## Environment Variables

These are set in the config package. To view take a look at the docs for that package:

    go doc github.com/freeformz/shh/config

## Building Debs on Heroku

```bash
heroku apps:create freeformz-build-shh --buildpack git://github.com/kr/heroku-buildpack-go.git
git push heroku
heroku open
```

Wait for the deb to be available, download and do what you want with it.

## 'Local' Development

1. Obtain a Linux system (only really tested on Ubuntu ATM)
2. Install Go v1.0.(2|3)
3. Set GOPATH [appropriately](http://golang.org/doc/code.html)
3. `go get github.com/freeformz/shh`
4. cd $GOPATH/src/github.com/freeformz/shh
5. go build
6. ./shh

hack away

## TODO

See [Github Issues](https://github.com/freeformz/shh/issues)
