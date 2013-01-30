package pollers

import (
	"bufio"
	"github.com/freeformz/shh/mm"
	"github.com/freeformz/shh/utils"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	TYPES_DEFAULTS = "btrfs,ext3,ext4,tmpfs,xfs"
)

var (
	typesEnv = utils.GetEnvWithDefault("SHH_DF_TYPES", TYPES_DEFAULTS)
	types    []string
)

func init() {
	types = strings.Split(typesEnv, ",")
	if !sort.StringsAreSorted(types) {
		sort.Strings(types)
	}
}

type Df struct {
	measurements chan<- *mm.Measurement
}

func NewDfPoller(measurements chan<- *mm.Measurement) Df {
	return Df{measurements: measurements}
}

func (poller Df) Poll(tick time.Time) {
	buf := new(syscall.Statfs_t)
	mountpoints := make(chan string)

	go feedMountpoints(mountpoints)

	for mp := range mountpoints {
		err := syscall.Statfs(mp, buf)
		if err != nil {
			log.Fatal(err)
		}
		mmp := massageMountPoint(mp)
		total_bytes := float64(uint64(buf.Bsize) * buf.Blocks)
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{mmp, "total_bytes"}, total_bytes}
		user_free_bytes := float64(uint64(buf.Bsize) * buf.Bavail)
		root_free_bytes := float64(uint64(buf.Bsize)*buf.Bfree) - user_free_bytes
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{mmp, "root", "free", "bytes"}, root_free_bytes}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{mmp, "user", "free", "bytes"}, user_free_bytes}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{mmp, "used", "bytes"}, total_bytes - root_free_bytes - user_free_bytes}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{mmp, "total", "inodes"}, float64(buf.Files)}
		poller.measurements <- &mm.Measurement{tick, poller.Name(), []string{mmp, "free", "inodes"}, float64(buf.Ffree)}
	}
}

func (poller Df) Name() string {
	return "df"
}

// Utility functions
func massageMountPoint(path string) string {
	switch path {
	case "/":
		return "root"
	}
	if strings.HasPrefix(path, "/") {
		path = strings.TrimLeft(path, "/")
	}
	path = strings.Replace(path, "/", "_", -1)
	return path
}

func feedMountpoints(mountpoints chan<- string) {
	defer close(mountpoints)
	file, err := os.Open("/proc/mounts")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fields := strings.Fields(line)
		fsType := fields[2]
		i := sort.SearchStrings(types, fsType)
		if i < len(types) && types[i] == fsType {
			mountpoints <- fields[1]
		}
	}
}
