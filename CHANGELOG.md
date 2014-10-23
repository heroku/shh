# Change Log

All notable changes to this project will be documented in this file.

## 0.6.4 - 2013-10-22

- Bugfix

## 0.6.3 - 2013-10-22

- SHH_LIBRATO_BATCH_SIZE defaults to 500
- SHH_LIBRATO_NETWORK_TIMEOUT defaults to 5s
- SHH_LIBRATO_BATCH_TIMEOUT defaults to SHH_INTERVAL
- Librato Outputter timeout doesn't start until there is a measurement

## 0.6.2 - 2013-10-22

- Report to librato how many guages / counters are being reported in a batch

## 0.6.1 - 2013-10-22

- `$SHH_LISTEN_TIMEOUT` now controls the timeout on the socket.
- Librato outlet now reports a `User-Agent` header at the request of Librato.
- Timeout errors to the librato api are now reported.
- Better handling of ntpdate sub process error messages.

## 0.6.0 - 2013-10-20

- shh-value cli tool for interacting with the unix socket.
- Latest version of Go (1.3.3) used.
- use github.com/heroku/slog for structured logging (extracted from shh originally).
- LISTEN Poller documentation.
- Improved Listen Poller with support for types and units
- Use of Go's logger (over fmt.Println)

## 0.5.0 - 2014-10-09

### Added

- Units. Measurements now have units which the Librato outputter takes
  advantage of. No other outputters currently take advantage of this.

### Changed (Breaking)

- sockstat poller now uses lowercase protocol names in emitted
  metrics. Previously, it broke convention and used
  uppercase. (i.e. 'UDP' is now 'udp')
- The percentage calculations for the "mem" and "df" pollers resulted
  in values between 0 and 1. CPU percentages are between 0-100. They
  now all use 0-100 instead of 0-1.

## Prior Versions

Sadly, we didn't keep a proper changelog for previous versions. :(
