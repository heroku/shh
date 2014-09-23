package main

import (
	"regexp"
	"time"
)

const (
	VERSION                 = "0.4.0"
	DEFAULT_INTERVAL        = "10s"                                                              // Default tick interval for pollers
	DEFAULT_OUTPUTTER       = "stdoutl2metder"                                                   // Default outputter
	DEFAULT_POLLERS         = "conntrack,cpu,df,disk,listen,load,mem,nif,ntpdate,processes,self" // Default pollers
	DEFAULT_PROFILE_PORT    = "0"                                                                // Default profile port, 0 disables
	DEFAULT_DF_TYPES        = "btrfs,ext3,ext4,tmpfs,xfs"                                        // Default fs types to report df for
	DEFAULT_NIF_DEVICES     = "eth0,lo"                                                          // Default interfaces to report stats for
	DEFAULT_CPU_AGGR        = false                                                              // Default whether to only report aggregate CPU
	DEFAULT_SYSLOGNG_SOCKET = "/var/lib/syslog-ng/syslog-ng.ctl"                                 // Default location of the syslog-ng socket
	DEFAULT_SOCKSTAT_PROTOS = "TCP,UDP,TCP6,UDP6"                                                // Default protocols to report sockstats on
	DEFAULT_PERCENTAGES     = ""                                                                 // Default pollers where publishing perc metrics is allowed
	DEFAULT_LIBRATO_URL     = "https://metrics-api.librato.com/v1/metrics"
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
	DfTypes               []string
	Listen                string
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
	SelfPollerMode        string
	Start                 time.Time
	DiskFilter            *regexp.Regexp
}

func GetConfig() (config Config) {
	config.Interval = GetEnvWithDefaultDuration("SHH_INTERVAL", DEFAULT_INTERVAL)    // Polling Interval
	config.Outputter = GetEnvWithDefault("SHH_OUTPUTTER", DEFAULT_OUTPUTTER)         // Outputter
	config.Pollers = GetEnvWithDefaultStrings("SHH_POLLERS", DEFAULT_POLLERS)        // Pollers to poll
	config.Source = GetEnvWithDefault("SHH_SOURCE", "")                              // Source to emit
	config.Prefix = GetEnvWithDefault("SHH_PREFIX", "")                              // Metric prefix to use
	config.ProfilePort = GetEnvWithDefault("SHH_PROFILE_PORT", DEFAULT_PROFILE_PORT) // Profile Port
	config.Percentages = GetEnvWithDefaultStrings("SHH_PERCENTAGES", DEFAULT_PERCENTAGES)
	config.DfTypes = GetEnvWithDefaultStrings("SHH_DF_TYPES", DEFAULT_DF_TYPES)                              // Default DF types
	config.Listen = GetEnvWithDefault("SHH_LISTEN", "unix,#shh")                                             // Default network socket info for listen
	config.NifDevices = GetEnvWithDefaultStrings("SHH_NIF_DEVICES", DEFAULT_NIF_DEVICES)                     // Devices to poll
	config.NtpdateServers = GetEnvWithDefaultStrings("SHH_NTPDATE_SERVERS", "0.pool.ntp.org,1.pool.ntp.org") // NTP Servers
	config.CpuOnlyAggregate = GetEnvWithDefaultBool("SHH_CPU_AGGR", false)                                   // Whether to only report aggregate CPU usage
	config.LibratoUrl = GetEnvWithDefault("SHH_LIBRATO_URL", DEFAULT_LIBRATO_URL)                            // The Librato API End-Point
	config.LibratoUser = GetEnvWithDefault("SHH_LIBRATO_USER", "")                                           // The Librato API User
	config.LibratoToken = GetEnvWithDefault("SHH_LIBRATO_TOKEN", "")                                         // The Librato API TOken
	config.LibratoBatchSize = GetEnvWithDefaultInt("SHH_LIBRATO_BATCH_SIZE", 200)                            // The max number of metrics to submit in a single request
	config.LibratoBatchTimeout = GetEnvWithDefaultDuration("SHH_LIBRATO_BATCH_TIMEOUT", "500ms")             // The max time metrics will sit un-delivered
	config.LibratoNetworkTimeout = GetEnvWithDefaultDuration("SHH_LIBRATO_NETWORK_TIMEOUT", "2s")            // The maximum time to wait for Librato to respond (for both dial and first header)
	config.CarbonHost = GetEnvWithDefault("SHH_CARBON_HOST", "")                                             // Where the Carbon Outputter sends it's data
	config.SockStatProtos = GetEnvWithDefaultStrings("SHH_SOCKSTAT_PROTOS", DEFAULT_SOCKSTAT_PROTOS)         // Protocols to report sockstats about
	config.StatsdHost = GetEnvWithDefault("SHH_STATSD_HOST", "")                                             // Where the Statsd Outputter sends it's data
	config.StatsdProto = GetEnvWithDefault("SHH_STATSD_PROTO", "udp")                                        // Whether the Stats Outputter uses TCP or UDP
	config.SyslogngSocket = GetEnvWithDefault("SHH_SYSLOGNG_SOCKET", DEFAULT_SYSLOGNG_SOCKET)                // The location of the syslog-ng socket
	config.SelfPollerMode = GetEnvWithDefault("SHH_SELF_POLLER_MODE", "full")                                // Self poller mode (full, minimal)
	tmp := GetEnvWithDefault("SHH_DISK_FILTER", ".*")
	config.DiskFilter = regexp.MustCompile(tmp)
	config.Start = start // Start time
	return
}
