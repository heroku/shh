Shh
----

System metric collection and reporting to STDOUT.

The general idea is that delivery should be seperated from collection.

Use something like [log-shuttle](https://github.com/ryandotsmith/log-shuttle) to deliver the metrics somewhere else.

This is mostly for me to learn some [Go](http://golang.org/).


TODO
-----

* Better types/interfaces for pollers
* Config? (do I really want a config)
* more collectors
