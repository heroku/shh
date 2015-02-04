package shh

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/heroku/shh/Godeps/_workspace/src/github.com/heroku/slog"
	redis "github.com/heroku/shh/Godeps/_workspace/src/github.com/xuyu/goredis"
)

var (
	// RedisKnownGauges lists the metrics that are known Gauges
	RedisKnownGauges = map[string]struct{}{
		"clients:connected_clients":          struct{}{},
		"clients:client_longest_output_list": struct{}{},
		"clients:client_biggest_input_buf":   struct{}{},
		"clients:blocked_clients":            struct{}{},

		"keyspace:db0.keys": struct{}{},

		"memory:used_memory":      struct{}{},
		"memory:used_memory_rss":  struct{}{},
		"memory:used_memory_peak": struct{}{},
		"memory:used_memory_lua":  struct{}{},

		"replication:master_last_io_seconds_ago":      struct{}{},
		"replication:master_sync_in_progress":         struct{}{},
		"replication:master_sync_left_bytes":          struct{}{},
		"replication:master_sync_last_io_seconds_ago": struct{}{},
		"replication:master_link_down_since_seconds":  struct{}{},
		"replication:connected_slaves":                struct{}{},

		"stats:instantaneous_ops_per_sec": struct{}{},
		"stats:pubsub_channels":           struct{}{},
		"stats:pubsub_patterns":           struct{}{},
		"stats:latest_fork_usec":          struct{}{},
	}
)

// Redis poller
// info contains a mapping for the section to the keys that you want for each key.
// See DEFAULT_REDIS_INFO comment
type Redis struct {
	measurements chan<- Measurement
	url          *url.URL
	info         map[string][]string
}

// NewRedisPoller constructs a functioning Redis poller from the provided config and reports on the provided channel
func NewRedisPoller(measurements chan<- Measurement, config Config) Redis {
	ctx := slog.Context{"poller": "redis", "fn": "NewRedisPoller"}
	info := make(map[string][]string)

	for _, sectionInfo := range strings.Split(config.RedisInfo, ";") {
		section := strings.Split(sectionInfo, ":")
		if len(section) == 2 {
			info[section[0]] = strings.Split(section[1], ",")
		} else {
			FatalError(ctx, fmt.Errorf("Expected sectionName:keys"), "")
		}
	}

	return Redis{measurements: measurements, url: config.RedisUrl, info: info}
}

// Poll executes the polling of the provided redis server.
func (poller Redis) Poll(tick time.Time) {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	cli, err := redis.DialURL(poller.url.String())
	if err != nil {
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error"}, 1, Errors}
		LogError(ctx, err, "connecting to redis")
		return
	}

	for section, sectionKeys := range poller.info {
		result, err := cli.Info(section)
		if err != nil {
			LogError(ctx, err, "for section "+section)
			poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"error", "info", section}, 1, Errors}
			continue
		}

		for _, line := range strings.Split(result, "\r\n") {
			key, rawValue := parseInfoLine(line)
			switch {
			case SliceContainsString(sectionKeys, key):
				poller.report(section, key, rawValue, tick)
			case strings.Contains(rawValue, "="):
				kvs := parseKeyValues(rawValue)

				for k, v := range kvs {
					subKey := key + "." + k
					if SliceContainsString(sectionKeys, subKey) {
						poller.report(section, subKey, v, tick)
					}
				}
			}
		}
	}

	defer cli.ClosePool()
}

func (poller Redis) report(section, subKey, rawValue string, tick time.Time) {
	value := Atouint64(rawValue)
	if _, ok := RedisKnownGauges[section+":"+subKey]; ok {
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{section, subKey}, value, Empty}
	} else {
		poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{section, subKey}, value, Empty}
	}
}

func parseInfoLine(line string) (key, value string) {
	bits := strings.Split(line, ":")
	if len(bits) == 2 {
		return bits[0], bits[1]
	}
	return "", ""
}

func parseKeyValues(values string) map[string]string {
	kvs := make(map[string]string)
	pairs := strings.Split(values, ",")
	for _, pair := range pairs {
		bits := strings.Split(pair, "=")
		if len(bits) == 2 {
			kvs[bits[0]] = bits[1]
		} else {
			LogError(slog.Context{"poller": "redis", "fn": "parseKeyValues"}, fmt.Errorf("Unexpected Format: len == %d", len(bits)), "")
		}
	}
	return kvs
}

// Name reports the name of this poller
func (poller Redis) Name() string {
	return "redis"
}

// Exit is a noop
func (poller Redis) Exit() {}
