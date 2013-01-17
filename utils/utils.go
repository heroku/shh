package utils

import (
	"strconv"
)

func Ui64toa(val uint64) string {
	return strconv.FormatUint(val, 10)
}
