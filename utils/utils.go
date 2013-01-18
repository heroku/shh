package utils

import (
	"os"
	"strconv"
)

func Ui64toa(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
