package shh

import(
  "encoding/json"
  "fmt"
  "net"
  "net/http"
  "time"

  "github.com/heroku/slog"
)

// {"atom_used":485956,"binary":1004525112,"code":11971352,"ets":4219429600}
type FolsomMemory struct {
  Total uint64 `json:"total"`
  Processes uint64 `json:"processes"`
  ProcessesUsed uint64 `json:"processes_used"`
  System uint64 `json:"system"`
  Atom uint64 `json:"atom"`
  AtomUsed uint64 `json:"atom_used"`
  Binary uint64 `json:"binary"`
  Code uint64 `json:"code"`
  Ets uint64 `json:"ets"`
}

type FolsomMetrics struct {
  Metrics []string `json:""`
}

type FolsomValue struct {
  Name string
  Value string `json:"value"`
}

type FolsomPoller struct {
  measurements chan<- Measurement
  baseUrl string
  client *http.Client
}

func NewFolsomPoller(measurements chan<- Measurement, config Config) FolsomPoller {
  var url string

  if config.FolsomBaseUrl != nil {
    url = config.FolsomBaseUrl.String()
  }

  client := &http.Client{
    Transport: &http.Transport{
      ResponseHeaderTimeout: config.NetworkTimeout,
      Dial: func(network, address string) (net.Conn, error) {
        return net.DialTimeout(network, address, config.NetworkTimeout)
      },
    },
  }

  return FolsomPoller{
    measurements: measurements,
		baseUrl:      url,
		client:       client,
  }
}

func (poller FolsomPoller) Poll(tick time.Time) {
	if poller.baseUrl == "" {
		return
	}

	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

  poller.doMemoryPoll(ctx, tick)
  // poller.doMetricsPoll(ctx, tick)
  // poller.doSystemPoll()

}

func (poller FolsomPoller) doMemoryPoll(ctx slog.Context, tick time.Time) () {
	resp, err := poller.doRequest("/_memory")
	if err != nil {
		LogError(ctx, err, "while performing request for this tick")
		return
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	memory := FolsomMemory{}
	if derr := decoder.Decode(&memory); derr != nil {
		LogError(ctx, derr, "while performing decode on response body")
		return
	}

	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "total"}, memory.Total, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "procs", "total"}, memory.Processes, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "procs", "used"}, memory.ProcessesUsed, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "system"}, memory.System, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "atom", "total"}, memory.Atom, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "atom", "used"}, memory.AtomUsed, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "binary"}, memory.Binary, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "code"}, memory.Code, Bytes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"mem", "ets"}, memory.Ets, Bytes}
}

func (poller FolsomPoller) doMetricsPoll(ctx slog.Context, tick time.Time) () {
}

func (poller FolsomPoller) doRequest(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", poller.baseUrl + path, nil)
	if err != nil {
		return nil, err
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

func (poller FolsomPoller) Name() string {
	return "erlfolsom"
}

func (poller FolsomPoller) Exit() {}
