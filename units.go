package shh

type Unit struct {
	name string
	abbr string
}

var (
	Empty        = Unit{"", ""}
	Percent      = Unit{"Percent", "%"}
	Bytes        = Unit{"Bytes", "b"}
	Seconds      = Unit{"Seconds", "s"}
	MilliSeconds = Unit{"MilliSeconds", "ms"}
	NanoSeconds  = Unit{"NanoSeconds", "ns"}
	Requests     = Unit{"Requests", "reqs"}
	Errors       = Unit{"Errors", "errs"}
	Packets      = Unit{"Packets", "pkts"}
	INodes       = Unit{"INodes", "inodes"}
	Files        = Unit{"Files", "files"}
	Processes    = Unit{"Processes", "procs"}
	Connections  = Unit{"Connections", "conns"}
	Sockets      = Unit{"Sockets", "socks"}
	Avg          = Unit{"Avg", "avg"}
	Objects      = Unit{"Objects", "objs"}
	Routines     = Unit{"Routines", "routines"}
)

func (u Unit) Name() string {
	return u.name
}

func (u Unit) Abbr() string {
	return u.abbr
}
