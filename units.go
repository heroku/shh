package main

type UnitType int

type Unit interface {
	Name() string
	Abbr() string
}

const (
	Empty UnitType = iota
	Percent
	Bytes
	Seconds
	MilliSeconds
	NanoSeconds
	Requests
	Errors
	Packets
	Ticks
	Avg
	INodes
	Files
	Processes
	Connections
	Sockets
)

var (
	UnitTypeNameMapping = map[UnitType]string{
		Empty:        "",
		Percent:      "Percent",
		Bytes:        "Bytes",
		Seconds:      "Seconds",
		MilliSeconds: "MilliSeconds",
		NanoSeconds:  "NanoSeconds",
		Requests:     "Requests",
		Errors:       "Errors",
		Packets:      "Packets",
		INodes:       "INodes",
		Files:        "Files",
		Processes:    "Processes",
		Connections:  "Connections",
		Sockets:      "Sockets",
	}
	UnitTypeAbbrMapping = map[UnitType]string{
		Empty:        "",
		Percent:      "%",
		Bytes:        "b",
		Seconds:      "s",
		MilliSeconds: "ms",
		NanoSeconds:  "ns",
		Requests:     "reqs",
		Errors:       "errs",
		Packets:      "pkts",
		Files:        "files",
		Processes:    "procs",
		Connections:  "conns",
		Sockets:      "socks",
	}
)

func (u UnitType) Name() string {
	return UnitTypeNameMapping[u]
}

func (u UnitType) Abbr() string {
	return UnitTypeAbbrMapping[u]
}
