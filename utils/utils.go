package utils

import (
	"log"
	"os"
	"strconv"
	"time"
)

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
	val, err := strconv.ParseUint(s, 10, 0)
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
