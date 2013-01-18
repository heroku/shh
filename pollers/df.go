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
)

const (
	TYPES_DEFAULTS = "btrfs,ext3,ext4,tmpfs,xfs"
)

var (
	typesEnv = os.Getenv("SHH_DF_TYPES")
	types    []string
)

func init() {
	// Nothing set, so defaults!
	if typesEnv == "" {
		typesEnv = TYPES_DEFAULTS
	}
	types = strings.Split(typesEnv, ",")
	if !sort.StringsAreSorted(types) {
		sort.Strings(types)
	}
}

type Df struct{}

func (poller Df) Poll(measurements chan<- *mm.Measurement) {
	buf := new(syscall.Statfs_t)
	mountpoints := make(chan string)

	go feedMountpoints(mountpoints)

	for mp := range mountpoints {
		err := syscall.Statfs(mp, buf)
		if err != nil {
			log.Fatal(err)
		}
		mmp := massageMountPoint(mp)
		measurements <- &mm.Measurement{poller.Name(), []string{mmp, "total_bytes"}, utils.Ui64toa(uint64(buf.Bsize) * buf.Blocks)}
		measurements <- &mm.Measurement{poller.Name(), []string{mmp, "free_bytes"}, utils.Ui64toa(uint64(buf.Bsize) * buf.Bfree)}
		measurements <- &mm.Measurement{poller.Name(), []string{mmp, "avail_bytes"}, utils.Ui64toa(uint64(buf.Bsize) * buf.Bavail)}
		measurements <- &mm.Measurement{poller.Name(), []string{mmp, "total_inodes"}, utils.Ui64toa(buf.Files)}
		measurements <- &mm.Measurement{poller.Name(), []string{mmp, "free_inodes"}, utils.Ui64toa(buf.Ffree)}
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
