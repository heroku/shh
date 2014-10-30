# Guide to Pollers in shh

## Pre-baked Pollers

`shh` ships with a large number of pollers which probably get you
pretty close to what you need.

### Conntrack (conntrack)

The conntrack poller produces 1 metric, which represents the total
number of open connections as reported by
`/proc/sys/net/netfilter/nf_conntrack_count`

The metric is emitted as: `<prefix>.conntrack.count`

### CPU (cpu)

`shh`'s built in CPU poller is based on the data found in
`/proc/stat`. The values expressed there are monotonically increasing,
resetting after every reboot. Since this information isn't very useful
for humans, it is converted into percentages by doing the following:

    totalDifference = sum(forall current[measure]) - sum(forall last[measure])
    this[measure] = (current[measure] - last[measure]) / totalDifference * 100

Where `this` is the reported measurement, `current` are the values
from `/proc/stat` now and `last` are the values reported on the last
poll. `current` replaces `last` for the next poll.

`shh` emits the following CPU metrics (as percentages between 0-100),
for each CPU:

* `<prefix>.cpu.user`
* `<prefix>.cpu.nice`
* `<prefix>.cpu.system`
* `<prefix>.cpu.idle`
* `<prefix>.cpu.iowait`
* `<prefix>.cpu.irq`
* `<prefix>.cpu.softirq`
* `<prefix>.cpu.steal`

For definitions of these metrics see the section on `/proc/stat` in
[man 5 proc][proc5].

### Disk Usage (df)

The `df` poller takes an interesting approach to determining which
mount points to report disk usage on. It is controlled by the
environment variable `SHH_DF_TYPES`, which should be provided as a
comma separated list of filesystem types (as used in `fstab`).

`/proc/mounts` is then read, and mounts that utilize one of the
configured filesystem types will report the following metrics in
bytes:

* `<prefix>.df.<mntpt>.total_bytes`
* `<prefix>.df.<mntpt>.root.free.bytes`
* `<prefix>.df.<mntpt>.user.free.bytes`
* `<prefix>.df.<mntpt>.used.bytes`

The poller will also include metrics about inodes:

* `<prefix>.df.<mntpt>.total.inodes`
* `<prefix>.df.<mntpt>.free.inodes`

And, if the environment variable `SHH_PERCENTAGES` includes `df`:

* `<prefix>.df.<mntpt>.used.perc`

which is the percentage used on the filesystem (0-1).

The `<mntpt>` here is a massaged version of the actual mount point,
which substitutes `_` for `/` to make the metric name better
compatible with graphing and collection tools.

### Disk IO Stats (disk)

Information about disk IO is collected by first reading from
[/proc][proc5] to get a list of partitions. Using this information it
gathers information from `/sys` which exposes the necessary
information to report Disk IO metrics:

* `<prefix>.disk.<device>.read.requests`
* `<prefix>.disk.<device>.read.merges`
* `<prefix>.disk.<device>.read.bytes`
* `<prefix>.disk.<device>.read.ticks`
* `<prefix>.disk.<device>.write.requests`
* `<prefix>.disk.<device>.write.merges`
* `<prefix>.disk.<device>.write.bytes`
* `<prefix>.disk.<device>.write.ticks`
* `<prefix>.disk.<device>.in_flight.requests`
* `<prefix>.disk.<device>.io.ticks`
* `<prefix>.disk.<device>.queue.time`

More information can be found in the [block stat][kernelstat]
documentation.

### Allocated/Open Files (file-nr)

The kernel provides information about the number of allocated file
structs (and given that there's 1 file per struct...), and the max
number it will allocate (which is customizable). This data is
presented to us through `/proc/sys/fs/file-nr`. The information is
presented in the following metrics:

* `<prefix>.filenr.alloc`
* `<prefix>.filenr.free`
* `<prefix>.filenr.max`

See [man 5 proc][proc5]

### Load Averages (load)

System load averages are given in the following metrics:

* `<prefix>.load.1m`
* `<prefix>.load.5m`
* `<prefix>.load.15m`

In addition, `/proc/loadavg` reports the number of currently runnable
processes/threads, and the total number of processes/threads that are
available to be executed:

* `<prefix>.load.scheduling.entities.executing`
* `<prefix>.load.scheduling.entities.total`

For completeness, `shh` also exposes the process id of the last process
started by the system:

* `<prefix>.load.pid.last`

### Memory (mem)

The `mem` poller uses `/proc/meminfo` which exposes a variable number of
measurements depending on the kernel version and configuration. By default,
`shh` will report a subset of these. Adding `mem` to `SHH__FULL` will tell
`shh` to report all of them. `shh` reports all measurements in bytes. See [man
5 proc][proc5] for more information on available data.

The template for the emitted metrics are:

* `<prefix>.mem.<fixup-name>`

where `<fixup-name>` is a lowercased version of the stat with '(' and ')'
replaced by '.', and '.' replaced by '_'.

In addition, if the environment variable `SHH_PERCENTAGES` includes `mem`
and/or `swap`:

* `<prefix>.memtotal.perc`
* `<prefix>.swaptotal.perc`

Which represents the total percentage of in use memory / swap (between 0-1).

### Network Interfaces (nif)

`shh` can report network interface status information as reported by
`/proc/net/dev`. To control which devices should be reported, use the
`SHH_NIF_DEVICES` environment variable, which should be a comma
separted list of network interfaces.

* `<prefix>.nif.<device>.receive.bytes`
* `<prefix>.nif.<device>.receive.packets`
* `<prefix>.nif.<device>.receive.errors`
* `<prefix>.nif.<device>.receive.dropped`
* `<prefix>.nif.<device>.receive.errors.fifo`
* `<prefix>.nif.<device>.receive.errors.frame`
* `<prefix>.nif.<device>.receive.compressed`
* `<prefix>.nif.<device>.receive.multicast`
* `<prefix>.nif.<device>.transmit.bytes`
* `<prefix>.nif.<device>.transmit.packets`
* `<prefix>.nif.<device>.transmit.errors`
* `<prefix>.nif.<device>.transmit.dropped`
* `<prefix>.nif.<device>.transmit.errors.fifo`
* `<prefix>.nif.<device>.transmit.errors.collisions`
* `<prefix>.nif.<device>.transmit.errors.carrier`
* `<prefix>.nif.<device>.transmit.compressed`

### NTP (ntpdate)

The ntpdate poller runs the command `ntpdate -q -u`. It reports:

* `<prefix>.ntpdate.offset.<server>`
* `<prefix>.ntpdate.delay.<server>`

The `<server>` value utilizes the values of the servers checked
against, which are configured via the `SHH_NTPDATE_SERVERS`
environment variable, and should be a fully-qualified domain name.

### Processes (processes)

The processes poller submits measurements of the count of processes in the
various process states. It uses `/proc/<pid>/stat` to get this information.

* `<prefix>.processes.running.count`
* `<prefix>.processes.sleeping.count`
* `<prefix>.processes.waiting.count`
* `<prefix>.processes.zombie.count`
* `<prefix>.processes.stopped.count`
* `<prefix>.processes.paging.count`

Additionally the processes poller will match the process names found in
`/proc/<pid>/stat` to the `SHH_PROCESSES_REGEX` and if the name mates it will
report these additional measurements:

* `<prefix>.processes.<process name>.procs.count`
* `<prefix>.processes.<process name>.threads.count`
* `<prefix>.processes.<process name>.cpu.sys.seconds`
* `<prefix>.processes.<process name>.cpu.user.seconds`
* `<prefix>.processes.<process name>.mem.pagefaults.minor.count`
* `<prefix>.processes.<process name>.mem.pagefaults.major.count`
* `<prefix>.processes.<process name>.mem.rss.bytes`
* `<prefix>.processes.<process name>.mem.stacksize.bytes`
* `<prefix>.processes.<process name>.mem.virtual.bytes`
* `<prefix>.processes.<process name>.io.read.bytes`
* `<prefix>.processes.<process name>.io.write.bytes`
* `<prefix>.processes.<process name>.io.read.ops`
* `<prefix>.processes.<process name>.io.write.ops`

### SHH self (self)

The self poller provides metrics by introspecting itself. The Go programming language makes this rather trivial through the [runtime][goruntime] package.

* `<prefix>.self.memstats.goroutines.num`
* `<prefix>.self.memstats.general.alloc`
* `<prefix>.self.memstats.general.alloc.bytes`
* `<prefix>.self.memstats.heap.alloc.bytes`
* `<prefix>.self.memstats.heap.inuse.bytes`

If the environment variable `SHH_FULL` contains "self", it also reports the following:

* `<prefix>.self.measurements.length`
* `<prefix>.self.memstats.general.sys.bytes`
* `<prefix>.self.memstats.general.pointer.lookups`
* `<prefix>.self.memstats.general.mallocs`
* `<prefix>.self.memstats.general.frees`
* `<prefix>.self.memstats.heap.sys.bytes`
* `<prefix>.self.memstats.heap.idle.bytes`
* `<prefix>.self.memstats.heap.released.bytes`
* `<prefix>.self.memstats.heap.objects`
* `<prefix>.self.memstats.stack.inuse`
* `<prefix>.self.memstats.stack.sys`
* `<prefix>.self.memstats.mspan.inuse`
* `<prefix>.self.memstats.mspan.sys`
* `<prefix>.self.memstats.mcache.inuse`
* `<prefix>.self.memstats.mcache.sys`
* `<prefix>.self.memstats.buckhash.sys`
* `<prefix>.self.memstats.gc.next`
* `<prefix>.self.memstats.gc.pause`
* `<prefix>.self.memstats.gc.num`

### Sockstat (sockstat)

The sockstat poller uses that socket statistics found in `/proc/net/sockstat` and `/proc/net/sockstat6`. The collected metrics can be controlled by the comma separated list of protocols specified in the environment variable `SHH_SOCKSTAT_PROTOS`.

* `<prefix>.sockstat.<protocol>.alloc`
* `<prefix>.sockstat.<protocol>.inuse`
* `<prefix>.sockstat.<protocol>.mem`
* `<prefix>.sockstat.<protocol>.orphan`
* `<prefix>.sockstat.<protocol>.tw`

## Writing your own poller

`shh` is written in the Go programming language, which doesn't support
dynamic linking. This makes building a plugin system fairly difficult,
but of course simplifies other aspects of the software lifecycle--most
notably, deployment.

However, there are 2 mechanisms which can be utilized to create your
own poller. The first involves using a builtin poller called "listen";
the other involves modifying the `shh` source code.

### The Listen Poller Approach

`shh` provides a facility for external processes to emit stats via a
socket. If you include the "listen" poller in `SHH_POLLERS`, `shh` will
create a listening socket, listening at the address described by
`SHH_LISTEN` (defaults to a unix socket called #shh in the CWD). If you
set `SHH_LISTEN_TIMEOUT` to a duration (defaults to the value of
`SHH_INTERVAL`) the socket will close if the timeout duration passed
without receiving any data.

Data is then communicated in the following format:

    <RFC3339 date stamp> <what> <value>\n
    
`<what>` is the metric name, and the interpretation of `<value>` is somewhat arbitrary:

The Poller will create a FloatGauge if the value parses as a floating
point number, and a counter otherwise.

The metrics will be emitted as:

    `<prefix>.listen.stats.<what>`
    
The listen poller also emits metrics about itself:

* `<prefix>.listen.stats.connection.count`
* `<prefix>.listen.stats.time.parse.errors`
* `<prefix>.listen.stats.value.parse.errors`
* `<prefix>.listen.stats.metrics`

#### Interpretation of SHH_LISTEN

The environment variable `SHH_LISTEN` is a comma separated value with 2 fields. The first field is the socket type (e.g. tcp, tcp4, tcp6, unix, unixpacket) and the second field is an appropriate address for that type, as specified by Go's [networking libraries][gonet].

### Make it first class

Making a first class poller is fairly simple. In this section we'll
develop a poller that, for every tick outputs a counter value equal to
1.

Pollers should be given a simple, but descriptive name. For our
example, we'll just call it `one`. We'll put the poller in
`one_poller.go`.

Though it'd be possible to make libraries of pollers and link them in
at compile time, the current strategy is to just put them all in
`package main`. We import `time` because the `Poll` method we'll
implement shortly takes one argument of type `time.Time`.

```go
package main

import (
  "time"
)
```

Pollers are just an interface that we need to implement. This block of
code implements or simple constant one counter:

```go
type One struct {
  measurements chan<- Measurement
}

func NewOnePoller(measurements chan<- Measurement) One {
  return One{measurements}
}

// Called on every tick from the main loop
func (poller One) Poll(tick time.Time) {
  poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{"one"}, 1}
}

func (poller One) Name() string {
  return "one"
}

// A finalizer of sorts
func (poller One) Exit() {}
```

Just implementing the Poller, however, doesn't automatically make it
available. Next, we must add our new Poller to `pollers.go` such that
we can make it available to the `MultiPoller`, which is how `shh`
internally collects metrics across many pollers at once.

Somewhere in the switch statement add the following:

```go
case "one":
  mp.RegisterPoller(NewOnePoller(measurements))
```

Adding "one" to the list of pollers in the environment variable
`SHH_POLLERS` will create a one poller, and start calling it's `Poll`
method for each tick.


[proc5]: http://linux.die.net/man/5/proc
[kernelstat]: http://www.kernel.org/doc/Documentation/block/stat.txt
[goruntime]: http://golang.org/pkg/runtime/#MemStats
[gonet]: http://golang.org/pkg/net/
