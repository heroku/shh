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
}

type Df struct{}

func massagePath(path string) string {
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

func (poller Df) Poll(measurements chan<- *mm.Measurement) {
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

		buf := new(syscall.Statfs_t)

		fields := strings.Fields(line)
		fsType := fields[2]
		i := sort.SearchStrings(types, fsType)
		if i < len(types) && types[i] == fsType {
			path := fields[1]
			err := syscall.Statfs(path, buf)
			if err != nil {
				log.Fatal(err)
			}
			massagedPath := massagePath(path)
			measurements <- &mm.Measurement{poller.Name(), []string{massagedPath, "total_bytes"}, utils.Ui64toa(uint64(buf.Bsize) * buf.Blocks)}
			measurements <- &mm.Measurement{poller.Name(), []string{massagedPath, "free_bytes"}, utils.Ui64toa(uint64(buf.Bsize) * buf.Bfree)}
			measurements <- &mm.Measurement{poller.Name(), []string{massagedPath, "avail_bytes"}, utils.Ui64toa(uint64(buf.Bsize) * buf.Bavail)}
			measurements <- &mm.Measurement{poller.Name(), []string{massagedPath, "total_inodes"}, utils.Ui64toa(buf.Files)}
			measurements <- &mm.Measurement{poller.Name(), []string{massagedPath, "free_inodes"}, utils.Ui64toa(buf.Ffree)}
		}
	}
}

func (poller Df) Name() string {
	return "df"
}
