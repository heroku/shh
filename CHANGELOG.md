# Change Log

All notable changes to this project will be documented in this file.

### 0.8.5 - 2013-10-31

- Fix reporting of per processes sys / user cpu

### 0.8.4 - 2013-10-30

- Ignore processes w/o names, likely due to a process exiting between
    enumerating the directory entries and reading /proc/<pid>/stat

### 0.8.3 - 2013-10-30

- Generate additional process stats for processes that match
    `SHH_PROCESSES_REGEX`.
- `SHH_PROCESSES_REGEX`: \A\z - Regex of process names to poll and extra
    additional measurements for
- `SHH_TICKS`: 100 - cpu ticks per second. Default should be correct for most
    systems. see `getconf CLK_TCK`. Temporary until we use cgo to get it
- `SHH_PAGE_SIZE`: 4096 - kernel page size. Default should be correct for most
    systems. See `getconf PAGESIZE`. Temporary until we use cgo to get it.

## 0.8.2 - 2013-10-28

- `SHH_LIBRATO_ROUND`: true - round measurement times to nearest interval
    during submission

## 0.8.1 - 2014-10-24

### Changed (Breaking)

- Some disk poller stats were incorrectly being reported as Gauges.

## 0.8.0 - 2014-10-22

### Added

- `SHH_DF_LOOP` introduced to avoid loopback mounts from showing up in df
    poller (default false)
- `SHH_FULL` introduced to add back full set of metrics from some pollers

### Changed

- Remove `DEFAULT_SELF_POLLER_MODE` in favor of SHH_FULL="self"
- `SHH_CPU_AGGR` defaults to true, eliminating cpu metrics for all cores by
    default
- `SHH_DF_TYPES` removes tmpfs by default
- Default pollers to minimal set. Utilize `SHH_FULL` to get full set of metrics

## 0.7.0 - 2014-10-22

- Update some defaults: `SHH_INTERVAL=60s` & `SHH_LIBRATO_BATCH_TIMEOUT=10s`

## 0.6.4 - 2014-10-22

- Bugfix

## 0.6.3 - 2014-10-22

- SHH_LIBRATO_BATCH_SIZE defaults to 500
- SHH_LIBRATO_NETWORK_TIMEOUT defaults to 5s
- SHH_LIBRATO_BATCH_TIMEOUT defaults to SHH_INTERVAL
- Librato Outputter timeout doesn't start until there is a measurement

## 0.6.2 - 2014-10-22

- Report to librato how many guages / counters are being reported in a batch

## 0.6.1 - 2014-10-22

- `$SHH_LISTEN_TIMEOUT` now controls the timeout on the socket.
- Librato outlet now reports a `User-Agent` header at the request of Librato.
- Timeout errors to the librato api are now reported.
- Better handling of ntpdate sub process error messages.

## 0.6.0 - 2014-10-20

- shh-value cli tool for interacting with the unix socket.
- Latest version of Go (1.3.3) used.
- use github.com/heroku/slog for structured logging (extracted from shh
    originally).
- LISTEN Poller documentation.
- Improved Listen Poller with support for types and units
- Use of Go's logger (over fmt.Println)

## 0.5.0 - 2014-10-09

### Added

- Units. Measurements now have units which the Librato outputter takes
    advantage of. No other outputters currently take advantage of this.

### Changed (Breaking)

- sockstat poller now uses lowercase protocol names in emitted metrics.
    Previously, it broke convention and used uppercase. (i.e. 'UDP' is now
    'udp')
- The percentage calculations for the "mem" and "df" pollers resulted in values
    between 0 and 1. CPU percentages are between 0-100. They now all use 0-100
    instead of 0-1.

## Prior Versions

Sadly, we didn't keep a proper changelog for previous versions. :(
