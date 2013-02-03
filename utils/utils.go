package utils

import (
	"bufio"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Fields(line string) []string {
	return strings.FieldsFunc(line, func(s rune) bool {
		switch s {
		case ':', ' ', '\n':
			return true
		}
		return false
	})
}

func SliceContainsString(ss []string, s string) bool {
	idx := sort.SearchStrings(ss, s)
	if idx < len(ss) && ss[idx] == s {
		return true
	}
	return false
}

func Ui64toa(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func Atofloat64(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal(err)
	}
	return val
}

func PercentFormat(val float64) string {
	return strconv.FormatFloat(val, 'f', 2, 64)
}

func Atouint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return val
}

// Checks to see if a path exists or not
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Returns the value of $env from the OS and if it's empty, returns def
func GetEnvWithDefault(env string, def string) string {
	tmp := os.Getenv(env)

	if tmp == "" {
		return def
	}

	return tmp
}

// Returns the value of $env from the OS and if it's empty, returns def
func GetEnvWithDefaultInt(env string, def int) int {
	tmp := os.Getenv(env)

	if tmp == "" {
		return def
	}

	i, err := strconv.Atoi(tmp)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func GetEnvWithDefaultDuration(env string, def string) time.Duration {
	tmp := os.Getenv(env)

	if tmp == "" {
		tmp = def
	}

	d, err := time.ParseDuration(tmp)

	if err != nil {
		log.Printf("$%s is not a valid duration\n", env)
		log.Fatal(err)
	}

	return d
}

// Returns a slice of sorted strings from the environment or default split on ,
// So "foo,bar" returns ["bar","foo"]
func GetEnvWithDefaultStrings(env string, def string) []string {
	tmp := strings.Split(GetEnvWithDefault(env, def), ",")
	if !sort.StringsAreSorted(tmp) {
		sort.Strings(tmp)
	}
	return tmp
}

// Returns a channeel that streams the lines from the file at fpath
func FileLineChannel(fpath string) <-chan string {
	c := make(chan string)

	go func(fpath string, cs chan<- string) {
		defer close(cs)
		file, err := os.Open(fpath)
		if err == nil {
			defer file.Close()
			buf := bufio.NewReader(file)

			for {
				line, err := buf.ReadString('\n')
				if err == nil {
					cs <- line
				} else {
					if err == io.EOF {
						break
					} else {
						log.Fatal(err)
					}
				}
			}
		}
	}(fpath, c)

	return c
}

func FixUpName(name string) []string {
	name = strings.ToLower(name)
	name = strings.Replace(name, "(", ".", -1)
	name = strings.Replace(name, ")", "", -1)
	name = strings.Replace(name, "_", ".", -1)
	return strings.Split(name, ".")
}
