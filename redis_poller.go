package shh

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/heroku/slog"
	redis "github.com/xuyu/goredis"
)

var (
	RedisKnownGauges = map[string]struct{}{
		"clients:connected_clients":                   struct{}{},
		"clients:client_longest_output_list":          struct{}{},
		"clients:client_biggest_input_buf":            struct{}{},
		"clients:blocked_clients":                     struct{}{},
		"stats:instantaneous_ops_per_sec":             struct{}{},
		"stats:pubsub_channels":                       struct{}{},
		"stats:pubsub_patterns":                       struct{}{},
		"stats:latest_fork_usec":                      struct{}{},
		"replication:master_last_io_seconds_ago":      struct{}{},
		"replication:master_sync_in_progress":         struct{}{},
		"replication:master_sync_left_bytes":          struct{}{},
		"replication:master_sync_last_io_seconds_ago": struct{}{},
		"replication:master_link_down_since_seconds":  struct{}{},
		"replication:connected_slaves":                struct{}{},
	}
)

type Redis struct {
	measurements chan<- Measurement
	url          *url.URL
	info         map[string][]string
}

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
			if SliceContainsString(sectionKeys, key) {
				value := Atouint64(rawValue)
				if _, ok := RedisKnownGauges[section+":"+key]; ok {
					poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{section, key}, value, Empty}
				} else {
					poller.measurements <- CounterMeasurement{tick, poller.Name(), []string{section, key}, value, Empty}
				}
			}
		}
	}

	defer cli.ClosePool()
}

func parseInfoLine(line string) (key, value string) {
	bits := strings.Split(line, ":")
	if len(bits) == 2 {
		return bits[0], bits[1]
	}
	return "", ""
}

func (poller Redis) Name() string {
	return "redis"
}

func (poller Redis) Exit() {}
