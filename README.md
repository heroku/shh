[![Travis](https://img.shields.io/travis/heroku/shh.svg)](https://travis-ci.org/heroku/shh)
[![Releases](https://img.shields.io/github/release/heroku/shh.svg)](https://github.com/heroku/shh/releases)
[![GoDoc](https://godoc.org/github.com/heroku/shh?status.svg)](http://godoc.org/github.com/heroku/shh)


System Heuristics Herald (aka Shh)
----

Gathers and relays system metrics

## Install

    go get github.com/heroku/shh

## Configuration

Configuration of shh doesn't use a config file, instead it uses environment variables.

| Environment Var | Type | Explanation | Default |
|:----------------|:-----|:------------|:--------|
| `SHH_INTERVAL` | duration | Polling Interval | 10s |
| `SHH_META` | bool | Report/Collect meta stats | false |
| `SHH_OUTPUTTER` | string | Outputter | stdoutl2metder |
| `SHH_POLLERS` | list of string | Pollers to poll | conntrack,cpu,df,disk,listen,load,mem,nif,ntpdate,processes,self |
| `SHH_SOURCE` | string | Source to emit | |
| `SHH_PREFIX` | string | Metric prefix to use | |
| `SHH_PROFILE_PORT` | string | Profile Port | 0 (off) |
| `SHH_PERCENTAGES` | list of string | Default pollers which should report percentages when applicable | |
| `SHH_DF_TYPES` | list of string | Default DF types | btrfs,ext3,ext4,tmpfs,xfs |
| `SHH_LISTEN` | string | Default network socket info for listen | unix,#shh |
| `SHH_LISTEN_TIMEOUT` | string | Socket timeout duration | `SHH_INTERVAL` |
| `SHH_NIF_DEVICES` | list of string | Devices to poll | eth0 |
| `SHH_NTPDATE_SERVERS` | list of string | NTP Servers | 0.pool.ntp.org,1.pool.ntp.org |
| `SHH_CPU_AGGR` | bool | Whether to only report aggregate CPU usage | true |
| `SHH_LIBRATO_USER` | string | The Librato API User | |
| `SHH_LIBRATO_TOKEN` | string | The Librato API Token | |
| `SHH_LIBRATO_URL` | string | The Librato API User | https://metrics-api.librato.com/v1/metrics |
| `SHH_LIBRATO_BATCH_SIZE` | int | The max number of metrics to submit in a single request | 500 |
| `SHH_LIBRATO_BATCH_TIMEOUT` | duration | The max time metrics will sit un-delivered | `SHH_INTERVAL` |
| `SHH_LIBRATO_ROUND` | bool | Should shh round times to the nearest interval? | true |
| `SHH_NETWORK_TIMEOUT` | duration | Timeout til connect (will retry). And timeout to first header (will assume successful). Used for HTTP(S) endpoints and other network communication | 5s |
| `SHH_CARBON_HOST` | string | Where the Carbon Outputter sends it's data | |
| `SHH_SOCKSTAT_PROTOS` | list of string | Protocols to report sockstats about | TCP,UDP,TCP6,UDP6 |
| `SHH_STATSD_HOST` | string | Where the Statsd Outputter sends it's data | |
| `SHH_STATSD_PROTO` | string | Whether the Stats Outputter uses TCP or UDP | udp |
| `SHH_SYSLOGNG_SOCKET` | string | The location of the syslog-ng socket | /var/lib/syslog-ng/syslog-ng.ctl |
| `SHH_FULL | list of strings | Pollers that should report full metrics. `shh` defaults to minimal | "" |
| `SHH_DISK_FILTER` | regexp | Scan devices that match this regex | (xv|s)d |
| `SHH_PROCESSES_REGEX` | regexp | Scan / extract metrics for processes that match this regex | \A\z |
| `SHH_TICKS` | int | cpu ticks per second: see `getconf CLK_TCK`. Default is probably correct. (temporary until we use cgo) | 100 |
| `SHH_PAGE_SIZE` | int | system page size in bytes: see `getconf PAGESIZE`. Default is probably correct. (temporary until we use cgo) | 4096 |
| `SHH_NAGIOS3_METRIC_NAMES` | list of strings | list of nagios 3 metric names to report stats on, see `nagios3stats -h` | NUMSERVICES,NUMHOSTS,AVGACTSVCLAT,AVGACTHSTLAT,NUMHSTACTCHK5M,NUMSVCACTCHK5M,NUMHSTACTCHK1M,NUMSVCACTCHK1M |
| `SHH_SPLUNK_PEERS_SKIP_VERIFY` | bool | whether or not to skip verification of HTTPS cert on splunk peers endpoint | false |
| `SHH_SPLUNK_PEERS_URL` | string | URL of splunk distributed peers status (e.g. https://user:pass@localhost:8089/services/search/distributed/peers?count=-1 | |
| `SHH_FOLSOM_BASE_URL` | string | URL of exported folsom metrics via folsome\_cowboy or folsom\_webmachine (e.g. https://localhost:5564/) | |
| `SHH_REDIS_URL` | string | URL for Redis as defined by [goredis](https://github.com/xuyu/goredis) (e.g. tcp://auth:password@127.0.0.1:6379/0?timeout=10s&maxidle=1) | tcp://localhost:6379/0?timeout=10s&maxidle=1 |
| `SHH_REDIS_INFO` | string | Description of [INFO](http://redis.io/commands/info): `section0:key0,key1;section1:key0,key1` to pull | clients:connected_clients;memory:used_memory,used_memory_rss;stats:instantaneous_ops_per_sec;keyspace:db0.keys |
| `SHH_CGROUPS` | list of string | cgroups to report stats on | group1,group2,group3 | empty (none) |


For more information on the duration type, see [time.ParseDuration](http://golang.org/pkg/time/#ParseDuration)

The regexp type supports valid regexps documented [here](http://golang.org/pkg/regexp/).

### A note about SHH_OUTPUTTER

The SHH_OUTPUTTER variable *may* not be enough on it's own to get the desired result. For instance, the Librato outputter, requires that `SHH_LIBRATO_USER` and `SHH_LIBRATO_TOKEN` be set.

### A note about SHH_PERCENTAGES

This variable works on "virtual" pollers and computes "percentage used", reporting as "<metric>.perc"

* mem (from the mem poller)
* swap (from the mem poller)
* df (from the df_poller)

## Building Debs

Requirements:

* dpkg (see also `brew install dpkg`)
* go & [gox](https://github.com/mitchellh/gox), which is installed via the Makefile

```bash
make debs
```

Note: You can find debs on the [Github release page](https://github.com/heroku/shh/releases)

## 'Local' Development

1. Obtain a Linux system (only really tested on Ubuntu ATM)
1. Install Go (version 1.4+)
1. Set GOPATH [appropriately](http://golang.org/doc/code.html)
1. `go get github.com/tools/godep`
1. `go get github.com/heroku/shh`
1. cd $GOPATH/src/github.com/heroku/shh
1. go test -v ./...

hack away

## Contributing

The goal for shh is to be a stable, low footprint system metrics
poller, and we welcome contributions, feedback and bug reports to make
that happen.

We're currently focused on supporting GNU/Linux systems, since that's
where we're using shh, but are open to supporting other platforms
provided the low footprint nature is preserved.

Please file bug reports through
[Github Issues](https://github.com/heroku/shh/issues). If you'd like
to contribute changes, please fork and submit a pull request.
