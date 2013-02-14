package config

import (
	"github.com/freeformz/shh/utils"
	"time"
)

const (
	VERSION              = "0.0.24"
	DEFAULT_INTERVAL     = "10s"                                                              // Default tick interval for pollers
	DEFAULT_OUTPUTTER    = "stdoutl2metder"                                                   // Default outputter
	DEFAULT_POLLERS      = "conntrack,cpu,df,disk,listen,load,mem,nif,ntpdate,processes,self" // Default pollers
	DEFAULT_PROFILE_PORT = "0"
	DEFAULT_DF_TYPES     = "btrfs,ext3,ext4,tmpfs,xfs"
	DEFAULT_NIF_DEVICES  = "eth0,lo"
)

var (
	Interval            = utils.GetEnvWithDefaultDuration("SHH_INTERVAL", DEFAULT_INTERVAL)                      // Polling Interval
	Outputter           = utils.GetEnvWithDefault("SHH_OUTPUTTER", DEFAULT_OUTPUTTER)                            // Outputter
	Pollers             = utils.GetEnvWithDefaultStrings("SHH_POLLERS", DEFAULT_POLLERS)                         // Pollers to poll
	Source              = utils.GetEnvWithDefault("SHH_SOURCE", "")                                              // Source to emit
	Prefix              = utils.GetEnvWithDefault("SHH_PREFIX", "")                                              // Metric prefix to use
	ProfilePort         = utils.GetEnvWithDefault("SHH_PROFILE_PORT", DEFAULT_PROFILE_PORT)                      // Profile Port
	DfTypes             = utils.GetEnvWithDefaultStrings("SHH_DF_TYPES", DEFAULT_DF_TYPES)                       // Default DF types
	Listen              = utils.GetEnvWithDefault("SHH_LISTEN", "unix,/tmp/shh")                                 // Default network socket info for listen
	NifDevices          = utils.GetEnvWithDefaultStrings("SHH_NIF_DEVICES", DEFAULT_NIF_DEVICES)                 // Devices to poll
	NtpdateServers      = utils.GetEnvWithDefaultStrings("SHH_NTPDATE_SERVERS", "0.pool.ntp.org,1.pool.ntp.org") // NTP Servers
	LibratoUser         = utils.GetEnvWithDefault("SHH_LIBRATO_USER", "")                                        // The Librato API User
	LibratoToken        = utils.GetEnvWithDefault("SHH_LIBRATO_TOKEN", "")                                       // The Librato API TOken
	LibratoBatchSize    = utils.GetEnvWithDefaultInt("SHH_LIBRATO_BATCH_SIZE", 50)                               // The max number of metrics to submit in a single request
	LibratoBatchTimeout = utils.GetEnvWithDefaultDuration("SHH_LIBRATO_BATCH_TIMEOUT", "500ms")                  // The max time metrics will sit un-delivered

	Start = time.Now() // Start time
)
