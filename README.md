System Heuristics Herald (aka Shh)
----

Gather and relay system metrics

## Install

    go get github.com/freeformz/shh

## environment variables

### General

    SHH_INTERVAL: The interval at which to poll. Defaults to "10s". See: http://golang.org/pkg/time/#ParseDuration
    SHH_SOURCE: The source for the metric if you want sources. No source is included if this isn't set.
    SHH_POLLERS: A comma seperated list of pollers to run. Defaults to "load,cpu,df,disk"

### Outputter related

    SHH_OUTPUTTER: The output module to use. Defaults to: "stdoutl2metder". Other choices are: "stdoutl2metraw" & "librato"
    SHH_LIBRATO_USER: When using the librato outputter, this is the librato username.
    SHH_LIBRATO_TOKEN: When using the librato outputter, this is the librato API token.
    SHH_LIBRATO_BATCH_SIZE: When using the librato outputter, this is the metric batch size for each POST. Defaults to: "50"
    SHH_LIBRATO_BATCH_TIMEOUT: When using the librato outputter, this is the timeout for a batch. Defaults to: "500ms"

### Poller related

    SHH_DF_TYPES: A comma seperated list of filesystem types (ext3, btrfs, tmpfs, etc) to return disk usage stats for. Defaults to: "btrfs,ext3,ext4,tmpfs,xfs"

## Building Debs on Heroku

```bash
heroku apps:create freeformz-build-shh --buildpack git://github.com/kr/heroku-buildpack-go.git
git push heroku
heroku open
```

Wait for the deb to be available, download and do what you want with it.

## TODO

* more collectors
    * conntrack connections
    * disk merged/octets/ops/time
    * memory buffered/cached/free/used
    * net tx/rx errors/octets/packets
    * processes blocked/fork_rate/paging/running/sleeping/stopped/zombies
    * swap cached/free/in/out/used
    * ntp stats
    * nagios 3 stats
    * process statistics
* small plugin interface for writing Exec'able plugins in any language
