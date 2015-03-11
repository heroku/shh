package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/heroku/shh"
)

var (
	versionFlag     = flag.Bool("version", false, "Display version info and exit")
	measurementType = flag.String("t", "g", "Measurement type gauge(g) or counter(c)")
	shhAddr         = flag.String("a", shh.DEFAULT_LISTEN_ADDR, "Address of a listening shh (protocol,addr)")
	unitFlag        = flag.String("u", "", "Unit of measurement and an optional abbreviation (ex. Bytes,b)")

	unitRegexp = regexp.MustCompile("[a-zA-Z$%#]+(,[a-zA-Z$%#]+)?")
)

func getConnection(addr string) (net.Conn, error) {
	bits := strings.Split(addr, ",")
	if len(bits) == 1 {
		return net.Dial("tcp", bits[0])
	}
	return net.Dial(bits[0], bits[1])
}

func die(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
	}
	flag.Usage()
	os.Exit(1)
}

func assertValidMetricName(mn string) string {
	if !shh.MetricNameRegexp.MatchString(mn) {
		die("ERROR: invalid metric name\n")
	}

	return mn
}

func assertValidUnit(unit string) string {
	if unit != "" && !unitRegexp.MatchString(unit) {
		die("ERROR: invalid unit\n")
	}

	return unit
}

func assertValidType(t string) string {
	if t == "gauge" || t == "g" {
		return "g"
	}
	if t == "counter" || t == "c" {
		return "c"
	}
	die("ERROR: invalid measurement type\n")
	return ""
}

func assertValidValue(v string) interface{} {
	vint, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		vflo, err := strconv.ParseFloat(v, 64)
		if err != nil {
			die("ERROR: invalid value\n")
		}
		return vflo
	}
	return vint
}

func formatLine(metric string, value interface{}, mtype string, unit string) string {
	ts := time.Now().Format(time.RFC3339)
	if unit == "" {
		return fmt.Sprintf("%s %s %v %s\n", ts, metric, value, mtype)
	}
	return fmt.Sprintf("%s %s %v %s %s\n", ts, metric, value, mtype, unit)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] <metric-name> <value>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *versionFlag {
		fmt.Println(shh.Version())
		os.Exit(0)
	}

	if flag.NArg() != 2 {
		die("")
	}

	metric := assertValidMetricName(flag.Arg(0))
	unit := assertValidUnit(*unitFlag)
	mmType := assertValidType(*measurementType)
	value := assertValidValue(flag.Arg(1))

	if !shh.MetricNameRegexp.MatchString(flag.Arg(0)) {
		die("ERROR: invalid metric name\n")
	}

	conn, err := getConnection(shh.GetEnvWithDefault("SHH_ADDRESS", *shhAddr))
	if err != nil {
		die(fmt.Sprintf("ERROR: couldn't get connection to %s: %s\n", *shhAddr, err))
	}
	conn.SetDeadline(time.Now().Add(time.Second * 5))
	line := formatLine(metric, value, mmType, unit)
	fmt.Printf(line)
	fmt.Fprintf(conn, line)
	conn.Close()
}
