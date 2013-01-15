Shh
----

System metric collection and reporting to STDOUT.

The general idea is that delivery should be seperated from collection.

Use something like [log-shuttle](https://github.com/ryandotsmith/log-shuttle) to deliver the metrics somewhere else.

This is mostly for me to learn some [Go](http://golang.org/).

## Install

    go get github.com/freeformz/shh

Building Debs on Heroku
---------------------------

```bash
heroku apps:create freeformz-build-shh --buildpack git://github.com/kr/heroku-buildpack-go.git
git push heroku
heroku open
```

Wait for the deb to be available, download and do what you want with it.

TODO
-----

* Better types/interfaces for pollers
* more collectors
    * conntrack connections
    * disk free/used
    * disk merged/octets/ops/time
    * memory buffered/cached/free/used
    * net tx/rx errors/octets/packets
    * processes blocked/fork_rate/paging/running/sleeping/stopped/zombies
    * swap cached/free/in/out/used
* small plugin interface for writing Exec'able plugins in any language
