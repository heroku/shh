package shh

import (
	"fmt"
	"github.com/freeformz/filechan"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	NonWord = regexp.MustCompile("\\W")
)

type Slog map[string]interface{}

func (s Slog) String() string {
	var sv string
	parts := make([]string, 0, len(s))

	for k, v := range s {
		switch v.(type) {
		case time.Time: // Format times the way we want them
			sv = v.(time.Time).Format(time.RFC3339Nano)
		default: // Let Go figure out the representation
			sv = fmt.Sprintf("%v", v)
		}
		// If there is a NonWord character then need to quote the value
		if NonWord.MatchString(sv) {
			sv = fmt.Sprintf("%q", sv)
		}
		// Assemble the final part and append it to the array
		parts = append(parts, fmt.Sprintf("%s=%s", k, sv))
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

func (s Slog) FatalError(err error, msg interface{}) {
	s.Error(err, msg)
	os.Exit(1)
}

func (s Slog) Error(err error, msg interface{}) {
	s["at"] = time.Now()
	s["error"] = err
	s["message"] = msg
	fmt.Println(s)
	delete(s, "error")
	delete(s, "message")
}

func Fields(line string) []string {
	var insideParens = false

	return strings.FieldsFunc(line, func(s rune) bool {
		switch s {
		case '(':
			insideParens = true
		case ')':
			insideParens = false
		case ';', ':', ' ', '\n':
			return insideParens == false
		}
		return false
	})
}

func LinearSliceContainsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func SliceContainsString(ss []string, s string) bool {
	if sort.StringsAreSorted(ss) {
		idx := sort.SearchStrings(ss, s)
		if idx < len(ss) && ss[idx] == s {
			return true
		}
	} else {
		return LinearSliceContainsString(ss, s)
	}
	return false
}

func Ui64toa(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func Atofloat64(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		Slog{"fn": "Atofloat64", "input": s}.FatalError(err, "converting string to float64")
	}
	return val
}

func PercentFormat(val float64) string {
	return strconv.FormatFloat(val, 'f', 2, 64)
}

func Atouint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		Slog{"fn": "Atouint64", "input": s}.FatalError(err, "converting string to uint64")
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
		Slog{"fn": "GetEnvWithDefaultInt", "env": env, "def": def}.FatalError(err, "converting atoi")
	}
	return i
}

// Returns the value of $env from the OS and if it's empty, returns def
func GetEnvWithDefaultBool(env string, def bool) bool {
	tmp := os.Getenv(env)

	if tmp == "" {
		return def
	}

	b, err := strconv.ParseBool(tmp)
	if err != nil {
		Slog{"fn": "GetEnvWithDefaultBool", "env": env, "def": def}.FatalError(err, "converting atob")
	}
	return b
}

func GetEnvWithDefaultDuration(env string, def string) time.Duration {
	tmp := os.Getenv(env)

	if tmp == "" {
		tmp = def
	}

	d, err := time.ParseDuration(tmp)

	if err != nil {
		Slog{"fn": "GetEnvWithDefaultDuration", "env": env, "def": def}.FatalError(err, "not a valid duration")
	}

	return d
}

// Returns a slice of sorted strings from the environment or default split on ,
// So "foo,bar" returns ["bar","foo"]
func GetEnvWithDefaultStrings(env string, def string) []string {
	env = GetEnvWithDefault(env, def)
	if env == "" {
		return make([]string, 0)
	}
	tmp := strings.Split(env, ",")
	if !sort.StringsAreSorted(tmp) {
		sort.Strings(tmp)
	}
	return tmp
}

// Small wrapper to handle errors on open
func FileLineChannel(fpath string) <-chan string {
	c, err := filechan.FileLineChannel(fpath)

	if err != nil {
		Slog{"fn": "FileLineChannel", "fpath": fpath}.FatalError(err, "creating FileLineChannel")
	}

	return c
}

func FixUpName(name string) []string {
	name = strings.ToLower(name)
	name = strings.Replace(name, "(", ".", -1)
	name = strings.Replace(name, ")", "", -1)
	name = strings.Replace(name, "_", ".", -1)
	name = strings.TrimLeft(name, ".")
	return strings.Split(name, ".")
}
