package shh

import (
	"fmt"
	"net/url"
	"regexp"
	"runtime"
	"time"
)

const (
	VERSION                          = "0.9.4"
	DEFAULT_EMPTY_STRING             = ""
	DEFAULT_INTERVAL                 = "60s"                                                              // Default tick interval for pollers
	DEFAULT_OUTPUTTER                = "stdoutl2metder"                                                   // Default outputter
	DEFAULT_POLLERS                  = "conntrack,cpu,df,disk,listen,load,mem,nif,ntpdate,processes,self" // Default pollers
	DEFAULT_PROFILE_PORT             = "0"                                                                // Default profile port, 0 disables
	DEFAULT_DF_TYPES                 = "btrfs,ext3,ext4,xfs"                                              // Default fs types to report df for
	DEFAULT_DF_LOOP                  = false                                                              // Default to not reporting df metrics for loop back filesystems
	DEFAULT_NIF_DEVICES              = "eth0,lo"                                                          // Default interfaces to report stats for
	DEFAULT_NTPDATE_SERVERS          = "0.pool.ntp.org,1.pool.ntp.org"                                    // Default to the pool.ntp.org servers
	DEFAULT_CPU_AGGR                 = true                                                               // Default whether to only report aggregate CPU
	DEFAULT_SYSLOGNG_SOCKET          = "/var/lib/syslog-ng/syslog-ng.ctl"                                 // Default location of the syslog-ng socket
	DEFAULT_SELF_POLLER_MODE         = "minimal"                                                          // Default to only minimal set of self metrics
	DEFAULT_SOCKSTAT_PROTOS          = "TCP,UDP,TCP6,UDP6"                                                // Default protocols to report sockstats on
	DEFAULT_PERCENTAGES              = ""                                                                 // Default pollers where publishing perc metrics is allowed
	DEFAULT_FULL                     = ""                                                                 // Default list of pollers who should report full metrycs
	DEFAULT_LIBRATO_URL              = "https://metrics-api.librato.com/v1/metrics"                       // Default librato url to submit metrics to
	DEFAULT_LIBRATO_BATCH_SIZE       = 500                                                                // Default submission count
	DEFAULT_LIBRATO_BATCH_TIMEOUT    = "10s"                                                              // Default submission after
	DEFAULT_LIBRATO_ROUND            = true                                                               // Round measure_time to interval
	DEFAULT_LISTEN_ADDR              = "unix,#shh"                                                        // listen on UDS #shh
	DEFAULT_DISK_FILTER              = "(xv|s)d"                                                          // xvd* and sd* by default
	DEFAULT_PROCESSES_REGEX          = `\A\z`                                                             // Regex of processes to pull additional stats about
	DEFAULT_TICKS                    = 100                                                                // Default number of clock ticks per second (see _SC_CLK_TCK)
	DEFAULT_PAGE_SIZE                = 4096                                                               // Default system page size (see getconf PAGESIZE)
	DEFAULT_NAGIOS3_METRIC_NAMES     = "NUMSERVICES,NUMHOSTS,AVGACTSVCLAT,AVGACTHSTLAT,NUMHSTACTCHK5M,NUMSVCACTCHK5M,NUMHSTACTCHK1M,NUMSVCACTCHK1M"
	DEFAULT_SPLUNK_PEERS_SKIP_VERIFY = false
	DEFAULT_NETWORK_TIMEOUT          = "5s"
	DEFAULT_REDIS_INFO               = "clients:connected_clients;memory:used_memory,used_memory_rss;stats:instantaneous_ops_per_sec"
	DEFAULT_REDIS_URL                = "tcp://localhost:6379/0?timeout=10s&maxidle=1"
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
	LibratoUrl            *url.URL
	LibratoUser           string
	LibratoToken          string
	LibratoBatchSize      int
	LibratoBatchTimeout   time.Duration
	LibratoRound          bool
	NetworkTimeout        time.Duration
	CarbonHost            string
	SockStatProtos        []string
	StatsdHost            string
	StatsdProto           string
	SyslogngSocket        string
	Start                 time.Time
	DiskFilter            *regexp.Regexp
	UserAgent             string
	ProcessesRegex        *regexp.Regexp
	Ticks                 int
	PageSize              int
	Nagios3MetricNames    []string
	SplunkPeersSkipVerify bool
	SplunkPeersUrl        *url.URL
	FolsomBaseUrl         *url.URL
	RedisUrl              *url.URL
	RedisInfo             string
}

func GetConfig() (config Config) {
	config.Interval = GetEnvWithDefaultDuration("SHH_INTERVAL", DEFAULT_INTERVAL)                                          // Polling Interval
	config.Outputter = GetEnvWithDefault("SHH_OUTPUTTER", DEFAULT_OUTPUTTER)                                               // Outputter
	config.Pollers = GetEnvWithDefaultStrings("SHH_POLLERS", DEFAULT_POLLERS)                                              // Pollers to poll
	config.Source = GetEnvWithDefault("SHH_SOURCE", DEFAULT_EMPTY_STRING)                                                  // Source to emit
	config.Prefix = GetEnvWithDefault("SHH_PREFIX", DEFAULT_EMPTY_STRING)                                                  // Metric prefix to use
	config.ProfilePort = GetEnvWithDefault("SHH_PROFILE_PORT", DEFAULT_PROFILE_PORT)                                       // Profile Port
	config.Percentages = GetEnvWithDefaultStrings("SHH_PERCENTAGES", DEFAULT_PERCENTAGES)                                  // Use Percentages for these pollers
	config.Full = GetEnvWithDefaultStrings("SHH_FULL", DEFAULT_FULL)                                                       // Report full measurements for these pollers
	config.DfTypes = GetEnvWithDefaultStrings("SHH_DF_TYPES", DEFAULT_DF_TYPES)                                            // Default DF types
	config.DfLoop = GetEnvWithDefaultBool("SHH_DF_LOOP", DEFAULT_DF_LOOP)                                                  // Report df metrics for loop back filesystmes or not
	config.Listen = GetEnvWithDefault("SHH_LISTEN", DEFAULT_LISTEN_ADDR)                                                   // Default network socket info for listen
	config.ListenTimeout = GetEnvWithDefaultDuration("SHH_LISTEN_TIMEOUT", config.Interval.String())                       // Listen Poller Socket Timeout
	config.NifDevices = GetEnvWithDefaultStrings("SHH_NIF_DEVICES", DEFAULT_NIF_DEVICES)                                   // Devices to poll
	config.NtpdateServers = GetEnvWithDefaultStrings("SHH_NTPDATE_SERVERS", DEFAULT_NTPDATE_SERVERS)                       // NTP Servers
	config.CpuOnlyAggregate = GetEnvWithDefaultBool("SHH_CPU_AGGR", DEFAULT_CPU_AGGR)                                      // Whether to only report aggregate CPU usage
	config.LibratoUrl = GetEnvWithDefaultURL("SHH_LIBRATO_URL", DEFAULT_LIBRATO_URL)                                       // The Librato API End-Point
	config.LibratoUser = GetEnvWithDefault("SHH_LIBRATO_USER", DEFAULT_EMPTY_STRING)                                       // The Librato API User
	config.LibratoToken = GetEnvWithDefault("SHH_LIBRATO_TOKEN", DEFAULT_EMPTY_STRING)                                     // The Librato API TOken
	config.LibratoBatchSize = GetEnvWithDefaultInt("SHH_LIBRATO_BATCH_SIZE", DEFAULT_LIBRATO_BATCH_SIZE)                   // The max number of metrics to submit in a single request
	config.LibratoBatchTimeout = GetEnvWithDefaultDuration("SHH_LIBRATO_BATCH_TIMEOUT", DEFAULT_LIBRATO_BATCH_TIMEOUT)     // The max time metrics will sit un-delivered
	config.LibratoRound = GetEnvWithDefaultBool("SHH_LIBRATO_ROUND", DEFAULT_LIBRATO_ROUND)                                // Should we round measurement times to the nearest Interval when submitting to Librato
	config.CarbonHost = GetEnvWithDefault("SHH_CARBON_HOST", DEFAULT_EMPTY_STRING)                                         // Where the Carbon Outputter sends it's data
	config.SockStatProtos = GetEnvWithDefaultStrings("SHH_SOCKSTAT_PROTOS", DEFAULT_SOCKSTAT_PROTOS)                       // Protocols to report sockstats about
	config.StatsdHost = GetEnvWithDefault("SHH_STATSD_HOST", DEFAULT_EMPTY_STRING)                                         // Where the Statsd Outputter sends it's data
	config.StatsdProto = GetEnvWithDefault("SHH_STATSD_PROTO", "udp")                                                      // Whether the Stats Outputter uses TCP or UDP
	config.SyslogngSocket = GetEnvWithDefault("SHH_SYSLOGNG_SOCKET", DEFAULT_SYSLOGNG_SOCKET)                              // The location of the syslog-ng socket
	config.ProcessesRegex = GetEnvWithDefaultRegexp("SHH_PROCESSES_REGEX", DEFAULT_PROCESSES_REGEX)                        // The regex to match process names against for collecting additional measurements
	config.Ticks = GetEnvWithDefaultInt("SHH_TICKS", DEFAULT_TICKS)                                                        // Number of ticks per CPU cycle. It's normally 100, but you can check with `getconf CLK_TCK`
	config.PageSize = GetEnvWithDefaultInt("SHH_PAGE_SIZE", DEFAULT_PAGE_SIZE)                                             // System Page Size. It's usually 4096, but you can check with `getconf PAGESIZE`
	config.Nagios3MetricNames = GetEnvWithDefaultStrings("SHH_NAGIOS3_METRIC_NAMES", DEFAULT_NAGIOS3_METRIC_NAMES)         // Which nagios3stats metrics names should we poll
	config.SplunkPeersSkipVerify = GetEnvWithDefaultBool("SHH_SPLUNK_PEERS_SKIP_VERIFY", DEFAULT_SPLUNK_PEERS_SKIP_VERIFY) // If SHH_SPLUNK_PEERS_URL is https, do we need to skip verification?
	config.SplunkPeersUrl = GetEnvWithDefaultURL("SHH_SPLUNK_PEERS_URL", DEFAULT_EMPTY_STRING)                             // URL of splunk search peers api endpoint. Ex: https://user:pass@localhost:8089/services/search/distributed/peers?count=-1
	config.FolsomBaseUrl = GetEnvWithDefaultURL("SHH_FOLSOM_BASE_URL", DEFAULT_EMPTY_STRING)                               // URL of splunk search peers api endpoint. Ex: https://user:pass@localhost:8089/services/search/distributed/peers?count=-1
	config.RedisUrl = GetEnvWithDefaultURL("SHH_REDIS_URL", DEFAULT_REDIS_URL)                                             // URL of redis endpoint. Ex: tcp://auth:pass@localhost:6379/0?timeout=10s&maxidle=1"
	config.RedisInfo = GetEnvWithDefault("SHH_REDIS_INFO", DEFAULT_REDIS_INFO)                                             // section:key1,key2;section2:key1,key2
	config.NetworkTimeout = GetEnvWithDefaultDuration("NETWORK_TIMEOUT", DEFAULT_NETWORK_TIMEOUT)                          // The maximum time to wait for network requests to respond (for both dial and first header when applicable)

	tmp := GetEnvWithDefault("SHH_DISK_FILTER", DEFAULT_DISK_FILTER)
	config.DiskFilter = regexp.MustCompile(tmp)
	config.UserAgent = fmt.Sprintf("shh/%s (%s; %s; %s; %s)", VERSION, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
	config.Start = start // Start time
	return
}
