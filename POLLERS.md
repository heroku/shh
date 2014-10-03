# Pre-baked Pollers

## Conntrack (conntrack)

The conntrack poller produces 1 metric, which represents the total
number of open connections as reported by
`/proc/sys/net/netfilter/nf_conntrack_count`

The metric is emitted as: `<prefix>.conntrack.count`

## CPU (cpu)

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
[proc(5)][proc5].

## Disk Usage (df)

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

## Disk IO Stats (disk)

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

## Allocated/Open Files (file-nr)

*TO BE WRITTEN*

## Load Averages (load)

*TO BE WRITTEN*

## Memory (mem)

*TO BE WRITTEN*

## Network Interfaces (nif)

*TO BE WRITTEN*

## NTP Date Skew (ntpdate)

*TO BE WRITTEN*

## Processes (processes)

*TO BE WRITTEN*

## Self (e.g. SHH) (self)

*TO BE WRITTEN*

## Sockstat (sockstat)

*TO BE WRITTEN*


# Writing your own poller

shh is written in the Go programming language, which doesn't support
dynamic linking. This makes building a plugin system fairly difficult,
but of course simplifies other aspects of the software lifecycle--most
notably, deployment.

However, there are 2 mechanisms which can be utilized to create your
own poller. The first involves using a builtin poller called "listen",
and the other involves modifying the shh source code.

## The Listen Poller Approach

*TO BE WRITTEN*

## Make it first class

*TO BE WRITTEN*

[proc5]: http://linux.die.net/man/5/proc proc(5)
[kernelstat]: http://www.kernel.org/doc/Documentation/block/stat.txt
