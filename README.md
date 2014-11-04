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
| `SHH_OUTPUTTER` | string | Outputter | stdoutl2metder |
| `SHH_POLLERS` | list of string | Pollers to poll | conntrack,cpu,df,disk,listen,load,mem,nif,ntpdate,processes,self |
| `SHH_SOURCE` | string | Source to emit | |
| `SHH_PREFIX` | string | Metric prefix to use | |
| `SHH_PROFILE_PORT` | string | Profile Port | 0 (off) |
| `SHH_PERCENTAGES` | list of string | Default pollers which should report percentages when applicable | |
| `SHH_DF_TYPES` | list of string | Default DF types | btrfs,ext3,ext4,tmpfs,xfs |
| `SHH_LISTEN` | string | Default network socket info for listen | unix,#shh |
| `SHH_LISTEN_TIMEOUT` | string | Socket timeout duration | `SHH_INTERVAL` |
| `SHH_NIF_DEVICES` | list of string | Devices to poll | eth0,lo |
| `SHH_NTPDATE_SERVERS` | list of string | NTP Servers | 0.pool.ntp.org,1.pool.ntp.org |
| `SHH_CPU_AGGR` | bool | Whether to only report aggregate CPU usage | true |
| `SHH_LIBRATO_USER` | string | The Librato API User | |
| `SHH_LIBRATO_TOKEN` | string | The Librato API Token | |
| `SHH_LIBRATO_URL` | string | The Librato API User | https://metrics-api.librato.com/v1/metrics |
| `SHH_LIBRATO_BATCH_SIZE` | int | The max number of metrics to submit in a single request | 500 |
| `SHH_LIBRATO_BATCH_TIMEOUT` | duration | The max time metrics will sit un-delivered | `SHH_INTERVAL` |
| `SHH_LIBRATO_NETWORK_TIMEOUT` | duration | Timeout til connect (will retry). And timeout to first header (will assume successful) | 5s |
| `SHH_LIBRATO_ROUND` | bool | Should shh round times to the nearest interval? | true |
| `SHH_NETWORK_TIMEOUT` | duration | Timeout til connect (will retry). And timeout to first header (will assume successful) | 5s |
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

For more information on the duration type, see [time.ParseDuration](http://golang.org/pkg/time/#ParseDuration)

The regexp type supports valid regexps documented [here](http://golang.org/pkg/regexp/).

### A note about SHH_OUTPUTTER

The SHH_OUTPUTTER variable *may* not be enough on it's own to get the desired result. For instance, the Librato outputter, requires that `SHH_LIBRATO_USER` and `SHH_LIBRATO_TOKEN` be set. 

### A note about SHH_PERCENTAGES

This variable works on "virtual" pollers and computes "percentage used", reporting as "<metric>.perc"

* mem (from the mem poller)
* swap (from the mem poller)
* df (from the df_poller)


## Building Debs on Heroku

```bash
heroku apps:create freeformz-build-shh --buildpack git://github.com/kr/heroku-buildpack-go.git
git push heroku
heroku open
```

Wait for the deb to be available, download and do what you want with it.

## 'Local' Development

1. Obtain a Linux system (only really tested on Ubuntu ATM)
2. Install Go (version 1.3 works fine)
3. Set GOPATH [appropriately](http://golang.org/doc/code.html)
4. `go get github.com/tool/godep`
3. `go get github.com/heroku/shh`
4. cd $GOPATH/src/github.com/heroku/shh
5. godep go install ./...
6. $GOPATH/bin/shh

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

## License

Copyright 2013 - 2014, Edward Muller, and contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
