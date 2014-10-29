package shh

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/heroku/slog"
)

const (
	PROC = "/proc"
)

type ProcInfo struct {
	name                              string
	pid                               int
	state                             string
	numProcs, numThreads              uint64
	cpuSys, cpuUser                   float64
	pagefaultsMajor, pagefaultsMinor  uint64
	rss, stacksize, vm                uint64
	diskOctetsRead, diskOctetsWritten uint64
	diskOpsRead, diskOpsWrite         uint64
}

type Procs struct {
	measurements chan<- Measurement
	regex        *regexp.Regexp
	ticks        float64
	pageSize     uint64
}

func NewProcessesPoller(measurements chan<- Measurement, config Config) Procs {
	return Procs{
		measurements: measurements,
		regex:        config.ProcessesRegex,
		ticks:        float64(config.Ticks),
		pageSize:     uint64(config.PageSize),
	}
}

func (poller Procs) Poll(tick time.Time) {
	ctx := slog.Context{"poller": poller.Name(), "fn": "Poll", "tick": tick}

	dir, err := os.Open(PROC)
	if err != nil {
		FatalError(ctx, err, "opening "+PROC)
	}

	defer dir.Close()

	dirs, err := dir.Readdirnames(0)
	if err != nil {
		FatalError(ctx, err, "reading dir names")
	}

	var running, sleeping, waiting, zombie, stopped, paging uint64

	processes := make(map[string]ProcInfo)

	for _, proc := range dirs {

		pid, err := strconv.Atoi(proc)

		// Skip anything that isn't an int or < 1
		if err != nil || pid < 1 {
			continue
		}

		pInfo := poller.GetProcInfo(pid)

		switch pInfo.state {
		case "R":
			running++
		case "S":
			sleeping++
		case "D":
			waiting++
		case "Z":
			zombie++
		case "T":
			stopped++
		case "W":
			paging++
		}

		if poller.regex.MatchString(pInfo.name) {
			proc := processes[pInfo.name]

			proc.numProcs += 1
			proc.numThreads += pInfo.numThreads
			proc.cpuSys += pInfo.cpuSys
			proc.cpuUser += pInfo.cpuUser
			proc.pagefaultsMajor += pInfo.pagefaultsMajor
			proc.pagefaultsMinor += pInfo.pagefaultsMinor
			proc.rss += pInfo.rss
			proc.stacksize += pInfo.stacksize
			proc.vm += pInfo.vm
			proc.diskOctetsRead += pInfo.diskOctetsRead
			proc.diskOctetsWritten += pInfo.diskOctetsWritten
			proc.diskOpsRead += pInfo.diskOpsRead
			proc.diskOpsWrite += pInfo.diskOpsWrite

			processes[pInfo.name] = proc
		}
	}

	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"running", "count"}, running, Processes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"sleeping", "count"}, sleeping, Processes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"waiting", "count"}, waiting, Processes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"zombie", "count"}, zombie, Processes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"stopped", "count"}, stopped, Processes}
	poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{"paging", "count"}, paging, Processes}

	for name, proc := range processes {
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "procs", "count"}, proc.numProcs, Processes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "threads", "count"}, proc.numThreads, Threads}
		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{name, "cpu", "sys", "seconds"}, proc.cpuSys, Seconds}
		poller.measurements <- FloatGaugeMeasurement{tick, poller.Name(), []string{name, "cpu", "sys", "seconds"}, proc.cpuUser, Seconds}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "mem", "pagefaults", "minor", "count"}, proc.pagefaultsMinor, Faults}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "mem", "pagefaults", "major", "count"}, proc.pagefaultsMajor, Faults}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "mem", "rss", "byts"}, proc.rss, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "mem", "stacksize", "bytes"}, proc.stacksize, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "mem", "virtual", "bytes"}, proc.vm, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "io", "read", "bytes"}, proc.diskOctetsRead, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "io", "write", "bytes"}, proc.diskOctetsWritten, Bytes}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "io", "read", "ops"}, proc.diskOpsRead, Ops}
		poller.measurements <- GaugeMeasurement{tick, poller.Name(), []string{name, "io", "write", "ops"}, proc.diskOpsWrite, Ops}
	}

}

func (poller Procs) Name() string {
	return "processes"
}

func (poller Procs) Exit() {}

func (poller Procs) GetProcInfo(pid int) (pInfo ProcInfo) {
	pInfo.pid = pid
	poller.ScanProcStat(&pInfo)
	poller.ScanProcStatus(&pInfo)
	poller.ScanProcIo(&pInfo)
	return
}

func (poller Procs) ScanProcStat(pInfo *ProcInfo) {
	statFile := fmt.Sprintf("%s/%d/stat", PROC, pInfo.pid)

	statData, err := ioutil.ReadFile(statFile)

	// Skip errors and return an empty string
	if err != nil {
		return
	}

	fields := Fields(string(statData))
	pInfo.name = strings.TrimSuffix(strings.TrimPrefix(fields[1], "("), ")")
	pInfo.state = fields[2]

	if len(fields) >= 13 {
		pInfo.pagefaultsMinor = Atouint64(fields[9]) + Atouint64(fields[10])
		pInfo.pagefaultsMajor = Atouint64(fields[11]) + Atouint64(fields[12])
	}

	if len(fields) >= 17 {
		pInfo.cpuUser = (Atofloat64(fields[13]) / poller.ticks) + (Atofloat64(fields[15]) / poller.ticks)
		pInfo.cpuSys = (Atofloat64(fields[14]) / poller.ticks) + (Atofloat64(fields[16]) / poller.ticks)
	}

	if len(fields) >= 20 {
		pInfo.numThreads = Atouint64(fields[19])
	}

	if len(fields) >= 24 {
		pInfo.rss = Atouint64(fields[23]) * poller.pageSize
		pInfo.vm = Atouint64(fields[22])
	}

	return
}

func (poller Procs) ScanProcIo(pInfo *ProcInfo) {
	ioFile := fmt.Sprintf("%s/%d/io", PROC, pInfo.pid)

	ioData, err := ioutil.ReadFile(ioFile)

	// Skip errors and return an empty string
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(ioData), "\n") {
		fields := Fields(line)
		if len(fields) >= 2 {
			switch fields[0] {
			case "rchar":
				pInfo.diskOctetsRead = Atouint64(fields[1])
			case "wchar":
				pInfo.diskOctetsWritten = Atouint64(fields[1])
			case "syscr":
				pInfo.diskOpsRead = Atouint64(fields[1])
			case "syscw":
				pInfo.diskOpsWrite = Atouint64(fields[1])
			}
		}
	}
	return
}

func (poller Procs) ScanProcStatus(pInfo *ProcInfo) {
	statusFile := fmt.Sprintf("%s/%d/status", PROC, pInfo.pid)

	statusData, err := ioutil.ReadFile(statusFile)

	// Skip errors and return an empty string
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(statusData), "\n") {
		fields := Fields(line)
		if len(fields) >= 2 {
			switch fields[0] {
			case "VmStk":
				pInfo.stacksize = Atouint64(fields[1]) * 1024

			}
		}
	}

	return
}
