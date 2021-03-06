package shh

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/heroku/slog"
)

type SplunkPeers struct {
	Entries []SplunkPeerEntry `xml:"entry"`
}

type SplunkPeerEntry struct {
	Title string          `xml:"title"`
	Keys  []SplunkPeerKey `xml:"content>dict>key"`
}

type SplunkPeerKey struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type SplunkSearchPeersPoller struct {
	measurements chan<- Measurement
	url          string
	credentials  *url.Userinfo
	client       *http.Client
}

func NewSplunkSearchPeersPoller(measurements chan<- Measurement, config Config) SplunkSearchPeersPoller {
	var creds *url.Userinfo
	var url string

	// Extract the creds from the URL
	if config.SplunkPeersUrl != nil {
		creds = config.SplunkPeersUrl.User
		config.SplunkPeersUrl.User = nil
		url = config.SplunkPeersUrl.String()
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: config.SplunkPeersSkipVerify},
			ResponseHeaderTimeout: config.NetworkTimeout,
			Dial: func(network, address string) (net.Conn, error) {
				return net.DialTimeout(network, address, config.NetworkTimeout)
			},
		},
	}

	return SplunkSearchPeersPoller{
		measurements: measurements,
		url:          url,
		credentials:  creds,
		client:       client,
	}
}

func (poller SplunkSearchPeersPoller) Poll(tick time.Time) {
	if poller.url == "" {
		return
	}

	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	resp, err := poller.doRequest()
	if err != nil {
		LogError(ctx, err, "while performing request for this tick")
		return
	}

	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)
	entries := SplunkPeers{}
	if xerr := decoder.Decode(&entries); xerr != nil {
		LogError(ctx, xerr, "while performing decode on response body")
		return
	}

	total := len(entries.Entries)
	stats := make(map[string]uint64)

	for _, entry := range entries.Entries {
		for _, key := range entry.Keys {
			if key.Name == "status" {
				k := "status:" + key.Value
				stats[k] += 1
			} else if key.Name == "replicationStatus" {
				k := "replicationStatus:" + key.Value
				stats[k] += 1
			}
		}
	}

	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"total"}, uint64(total), Peers}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"up"}, stats["status:Up"], Peers}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"down"}, stats["status:Down"], Peers}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"replication", "success"}, stats["replicationStatus:Successful"], Peers}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"replication", "failed"}, stats["replicationStatus:Failed"], Peers}
}

func (poller SplunkSearchPeersPoller) doRequest() (*http.Response, error) {
	req, err := http.NewRequest("GET", poller.url, nil)
	if err != nil {
		return nil, err
	}

	if poller.credentials != nil {
		password, _ := poller.credentials.Password()
		req.SetBasicAuth(poller.credentials.Username(), password)
	}

	resp, rerr := poller.client.Do(req)
	if rerr != nil {
		return nil, rerr
	} else if resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("Response returned a %d", resp.StatusCode)
	}

	return resp, nil
}

func (poller SplunkSearchPeersPoller) Name() string {
	return "splunksearchpeers"
}

func (poller SplunkSearchPeersPoller) Exit() {}
