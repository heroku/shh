package shh

import (
	"fmt"
	"regexp"
	"runtime"
	"time"
)

const (
	VERSION                         = "0.8.0"
	DEFAULT_EMPTY_STRING            = ""
	DEFAULT_INTERVAL                = "60s"                                                              // Default tick interval for pollers
	DEFAULT_OUTPUTTER               = "stdoutl2metder"                                                   // Default outputter
	DEFAULT_POLLERS                 = "conntrack,cpu,df,disk,listen,load,mem,nif,ntpdate,processes,self" // Default pollers
	DEFAULT_PROFILE_PORT            = "0"                                                                // Default profile port, 0 disables
	DEFAULT_DF_TYPES                = "btrfs,ext3,ext4,xfs"                                              // Default fs types to report df for
	DEFAULT_DF_LOOP                 = false                                                              // Default to not reporting df metrics for loop back filesystems
	DEFAULT_NIF_DEVICES             = "eth0,lo"                                                          // Default interfaces to report stats for
	DEFAULT_NTPDATE_SERVERS         = "0.pool.ntp.org,1.pool.ntp.org"                                    // Default to the pool.ntp.org servers
	DEFAULT_CPU_AGGR                = true                                                               // Default whether to only report aggregate CPU
	DEFAULT_SYSLOGNG_SOCKET         = "/var/lib/syslog-ng/syslog-ng.ctl"                                 // Default location of the syslog-ng socket
	DEFAULT_SELF_POLLER_MODE        = "minimal"                                                          // Default to only minimal set of self metrics
	DEFAULT_SOCKSTAT_PROTOS         = "TCP,UDP,TCP6,UDP6"                                                // Default protocols to report sockstats on
	DEFAULT_PERCENTAGES             = ""                                                                 // Default pollers where publishing perc metrics is allowed
	DEFAULT_FULL                    = ""                                                                 // Default list of pollers who should report full metrycs
	DEFAULT_LIBRATO_URL             = "https://metrics-api.librato.com/v1/metrics"
	DEFAULT_LIBRATO_BATCH_SIZE      = 500
	DEFAULT_LIBRATO_NETWORK_TIMEOUT = "5s"
	DEFAULT_LIBRATO_BATCH_TIMEOUT   = "10s"
	DEFAULT_LISTEN_ADDR             = "unix,#shh"
	DEFAULT_DISK_FILTER             = "(xv|s)d"
)

var (
	start = time.Now()
)

type Config struct {
	Interval              time.Duration
	Outputter             string
	Pollers               []string
	Source                string
	Prefix                string
	ProfilePort           string
	Percentages           []string
	Full                  []string
	DfTypes               []string
	DfLoop                bool
	Listen                string
	ListenTimeout         time.Duration
	NifDevices            []string
	NtpdateServers        []string
	CpuOnlyAggregate      bool
	LibratoUrl            string
	LibratoUser           string
	LibratoToken          string
	LibratoBatchSize      int
	LibratoBatchTimeout   time.Duration
	LibratoNetworkTimeout time.Duration
	CarbonHost            string
	SockStatProtos        []string
	StatsdHost            string
	StatsdProto           string
	SyslogngSocket        string
	Start                 time.Time
	DiskFilter            *regexp.Regexp
	UserAgent             string
}

func GetConfig() (config Config) {
	config.Interval = GetEnvWithDefaultDuration("SHH_INTERVAL", DEFAULT_INTERVAL)                                            // Polling Interval
	config.Outputter = GetEnvWithDefault("SHH_OUTPUTTER", DEFAULT_OUTPUTTER)                                                 // Outputter
	config.Pollers = GetEnvWithDefaultStrings("SHH_POLLERS", DEFAULT_POLLERS)                                                // Pollers to poll
	config.Source = GetEnvWithDefault("SHH_SOURCE", DEFAULT_EMPTY_STRING)                                                    // Source to emit
	config.Prefix = GetEnvWithDefault("SHH_PREFIX", DEFAULT_EMPTY_STRING)                                                    // Metric prefix to use
	config.ProfilePort = GetEnvWithDefault("SHH_PROFILE_PORT", DEFAULT_PROFILE_PORT)                                         // Profile Port
	config.Percentages = GetEnvWithDefaultStrings("SHH_PERCENTAGES", DEFAULT_PERCENTAGES)                                    // Use Percentages for these pollers
	config.Full = GetEnvWithDefaultStrings("SHH_FULL", DEFAULT_FULL)                                                         // Report full measurements for these pollers
	config.DfTypes = GetEnvWithDefaultStrings("SHH_DF_TYPES", DEFAULT_DF_TYPES)                                              // Default DF types
	config.DfLoop = GetEnvWithDefaultBool("SHH_DF_LOOP", DEFAULT_DF_LOOP)                                                    // Report df metrics for loop back filesystmes or not
	config.Listen = GetEnvWithDefault("SHH_LISTEN", DEFAULT_LISTEN_ADDR)                                                     // Default network socket info for listen
	config.ListenTimeout = GetEnvWithDefaultDuration("SHH_LISTEN_TIMEOUT", config.Interval.String())                         // Listen Poller Socket Timeout
	config.NifDevices = GetEnvWithDefaultStrings("SHH_NIF_DEVICES", DEFAULT_NIF_DEVICES)                                     // Devices to poll
	config.NtpdateServers = GetEnvWithDefaultStrings("SHH_NTPDATE_SERVERS", DEFAULT_NTPDATE_SERVERS)                         // NTP Servers
	config.CpuOnlyAggregate = GetEnvWithDefaultBool("SHH_CPU_AGGR", DEFAULT_CPU_AGGR)                                        // Whether to only report aggregate CPU usage
	config.LibratoUrl = GetEnvWithDefault("SHH_LIBRATO_URL", DEFAULT_LIBRATO_URL)                                            // The Librato API End-Point
	config.LibratoUser = GetEnvWithDefault("SHH_LIBRATO_USER", DEFAULT_EMPTY_STRING)                                         // The Librato API User
	config.LibratoToken = GetEnvWithDefault("SHH_LIBRATO_TOKEN", DEFAULT_EMPTY_STRING)                                       // The Librato API TOken
	config.LibratoBatchSize = GetEnvWithDefaultInt("SHH_LIBRATO_BATCH_SIZE", DEFAULT_LIBRATO_BATCH_SIZE)                     // The max number of metrics to submit in a single request
	config.LibratoBatchTimeout = GetEnvWithDefaultDuration("SHH_LIBRATO_BATCH_TIMEOUT", DEFAULT_LIBRATO_BATCH_TIMEOUT)       // The max time metrics will sit un-delivered
	config.LibratoNetworkTimeout = GetEnvWithDefaultDuration("SHH_LIBRATO_NETWORK_TIMEOUT", DEFAULT_LIBRATO_NETWORK_TIMEOUT) // The maximum time to wait for Librato to respond (for both dial and first header)
	config.CarbonHost = GetEnvWithDefault("SHH_CARBON_HOST", DEFAULT_EMPTY_STRING)                                           // Where the Carbon Outputter sends it's data
	config.SockStatProtos = GetEnvWithDefaultStrings("SHH_SOCKSTAT_PROTOS", DEFAULT_SOCKSTAT_PROTOS)                         // Protocols to report sockstats about
	config.StatsdHost = GetEnvWithDefault("SHH_STATSD_HOST", DEFAULT_EMPTY_STRING)                                           // Where the Statsd Outputter sends it's data
	config.StatsdProto = GetEnvWithDefault("SHH_STATSD_PROTO", "udp")                                                        // Whether the Stats Outputter uses TCP or UDP
	config.SyslogngSocket = GetEnvWithDefault("SHH_SYSLOGNG_SOCKET", DEFAULT_SYSLOGNG_SOCKET)                                // The location of the syslog-ng socket
	tmp := GetEnvWithDefault("SHH_DISK_FILTER", DEFAULT_DISK_FILTER)
	config.DiskFilter = regexp.MustCompile(tmp)
	config.UserAgent = fmt.Sprintf("shh/%s (%s; %s; %s; %s)", VERSION, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
	config.Start = start // Start time
	return
}
