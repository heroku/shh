# Listen Poller

## Overview

The listen poller is a poller that allows external sources to record
metrics through a single shh instance. This means that any process
that can produce output, can record metrics through shh. While this is
happening, shh can concurrently collect and publish it's own metrics
(from other pollers).

## History

The listen poller was originally written such that a command like the
following:

    (while true; do 
        echo $(date "+%Y-%m-%dT%H:%M:%SZ") memfree \
             $(grep MemFree /proc/meminfo | awk '{print $2}').0; 
        sleep 5; 
    done) | nc -U \#shh
  
would send the string `2014-01-29T01:01:01Z memfree 29102.0\n` to the
Unix socket named #shh every 5 seconds. This corresponds to a RFC-3339
date, followed by metric name, followed by the value.

## Problem

`shh` now supports, in a first class manner, both units (e.g. bytes)
and specific data types such as counters, and gauges. shh has always
supported these types, but only via inference on the value
itself. Previously, integral values were considered counters, where as
floating point values were considered gauges.

In practice, this worked fine internally, but for externally written
pollers, it could potentially lead to problems, if slightly careless,
for outputters such as Librato where a counter and gauge are treated
differently, and writing a counter value to a gauge (and vice versa)
results in an error.

Since we'd like to start using the listen interface for pollers which
are complicated to write, or already require an external process to
gather the data (such as anything involving
[BPF](https://en.wikipedia.org/wiki/Berkeley_Packet_Filter) like `ss`
and friends), some care needs to be taken to ensure we're providing
an easy to use interface that's as compatible as possible to what the
builtin pollers can produce.

## Listen's Line format

The original line format was too simple to support this, and moving
forward we'd like to continue to keep it simple. We'd also like to do
so in a backwards compatible way. Thus, if a program provides shh with
nothing more than:

    2014-01-29T01:01:01Z memfree 29102.0\n
   
it should still do the "right thing." We, however, have now
implemented the following EBNF for unambiguously describing a metric:
  
    METRIC := DATE <SP> NAME <SP> VALUE <NL>
    
    DATE := <RFC-3339 DATE> | <UNIX TS>
    
    NAME := [a-zA-Z0-9]([a-zA-Z0-9.-]+)?
 
    VALUE := <FLOAT> META? |
             <INTEGER> META?
             
    META := `|` TYPE `:` UNIT |
            `|` TYPE
    
    UNIT := [a-zA-Z]+ |
            [a-zA-Z]+ `,` [a-zA-Z]+
            
    TYPE := `c` | `g`
    
The `UNIT` non-terminal describes the unit that the measurement is in,
with an optional abbreviation, *e.g.*, "Bytes,b" or "Seconds,s".

### Command Line Interface

While the simplicity of using a shell to execute:

    (while true; do 
        echo $(date "+%Y-%m-%dT%H:%M:%SZ") memfree \
             $(grep MemFree /proc/meminfo | awk '{print $2}').0;
        sleep 5; 
    done) | nc -U \#shh`
    
can't really be denied, a program that packages up and ships a metric
is nicer and certainly more ideal. We therefore created a new command
`shh-value`:

The job of `shh-value` is to properly format a metric and value that
it's given and post it to the socket given. `shh-value` has the
following usage:

    usage: shh-value [options] <metric-name> <value>
    
       -a ADDR     ADDR to connect to shh on (ex: unix,#shh)
       -h          this help message
       -t TYPE     TYPE is gauge (default) or counter
       -u UNIT     UNIT that measurement is in (ex: Bytes,b)
       -version    Version
       
    ADDR can also be set by the environment variable SHH_ADDRESS,
    which looks like <protocol>,<address>. See the Go documentation
    for details.
  
When invoked with the proper arguments, a connection will be made to
the given address and a measurement will be posted with the current
timestamp, and arguments:

    shh-value -a tcp,127.0.0.1:8000 -t gauge -u Bytes,b memfree 1093293
    
will open a TCP connection to 127.0.0.1:8000 and send the following:

    2014-01-29T01:01:01Z memfree 1093293|g:Bytes,b
