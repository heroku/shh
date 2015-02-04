package shh

import (
	"log"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/heroku/shh/Godeps/_workspace/src/github.com/freeformz/filechan"
	"github.com/heroku/shh/Godeps/_workspace/src/github.com/heroku/slog"
)

var (
	NonWord = regexp.MustCompile("\\W")
)

func FatalError(ctx slog.Context, err error, msg interface{}) {
	ctx["error"] = err
	ctx["message"] = msg
	ErrLogger.Fatal(ctx)
}

func LogError(ctx slog.Context, err error, msg interface{}) {
	ctx["error"] = err
	ctx["message"] = msg
	ErrLogger.Println(ctx)
	delete(ctx, "error")
	delete(ctx, "message")
}

func Fields(line string) []string {
	var insideParens = false

	return strings.FieldsFunc(line, func(s rune) bool {
		switch s {
		case '(':
			insideParens = true
		case ')':
			insideParens = false
		case ';', ':', ' ', '\t', '\n':
			return insideParens == false
		}
		return false
	})
}

func SliceContainsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func Ui64toa(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func Atofloat64(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		FatalError(slog.Context{"fn": "Atofloat64", "input": s}, err, "converting string to float64")
	}
	return val
}

func PercentFormat(val float64) string {
	return strconv.FormatFloat(val, 'f', 2, 64)
}

func Atouint64(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		FatalError(slog.Context{"fn": "Atouint64", "input": s}, err, "converting string to uint64")
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
		FatalError(slog.Context{"fn": "GetEnvWithDefaultInt", "env": env, "def": def}, err, "converting atoi")
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
		FatalError(slog.Context{"fn": "GetEnvWithDefaultBool", "env": env, "def": def}, err, "converting atob")
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
		FatalError(slog.Context{"fn": "GetEnvWithDefaultDuration", "env": env, "def": def}, err, "not a valid duration")
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

// Returns a *regexp.Regexp compiled from the env or default
func GetEnvWithDefaultRegexp(env string, def string) *regexp.Regexp {
	env = GetEnvWithDefault(env, def)
	re, err := regexp.Compile(env)
	if err != nil {
		FatalError(slog.Context{"fn": "GetEnvWithDefaultRegexp", "env": env, "def": def}, err, "not a valid regex")
	}
	return re
}

// Returns a *url.URL representing the given env, or nil if empty
func GetEnvWithDefaultURL(env string, def string) *url.URL {
	env = GetEnvWithDefault(env, def)
	if env == "" {
		return nil
	}

	parsed, err := url.Parse(env)
	if err != nil {
		FatalError(slog.Context{"fn": "GetEnvWithDefaultURL", "env": env, "def": def}, err, "not a valid URL")
	}
	return parsed
}

// Small wrapper to handle errors on open
func FileLineChannel(fpath string) <-chan string {
	c, err := filechan.FileLineChannel(fpath)

	if err != nil {
		FatalError(slog.Context{"fn": "FileLineChannel", "fpath": fpath}, err, "creating FileLineChannel")
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
