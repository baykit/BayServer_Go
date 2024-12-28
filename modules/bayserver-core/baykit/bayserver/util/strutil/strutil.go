package strutil

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"golang.org/x/text/encoding/htmlindex"
	"slices"
	"strconv"
	"strings"
)

var falses = []string{"no", "false", "0", "off"}
var trues = []string{"yes", "true", "1", "on"}

func ParseBool(val string) (bool, exception.Exception) {
	val = strings.ToLower(val)
	if slices.Contains(trues, val) {
		return true, nil

	} else if slices.Contains(falses, val) {
		return false, nil

	} else {
		return false, exception.NewException("Invalid boolean value: %s", val)
	}
}

func ParseInt(val string) (int, exception.Exception) {
	ival, err := strconv.Atoi(val)
	if err != nil {
		return 0, exception.NewException("Invalid integer value: %s", val)
	}
	return ival, nil
}

func ParseSize(val string) (int, exception.Exception) {
	val = strings.ToLower(val)
	var rate = 1
	var err exception.Exception = nil
	if EndsWith(val, "b") {
		val = val[0 : len(val)-1]

	} else if EndsWith(val, "k") {
		val = val[0 : len(val)-1]
		rate = 1024

	} else if EndsWith(val, "m") {
		val = val[0 : len(val)-1]
		rate = 1024 * 1024
	}

	ival, ex := strconv.Atoi(val)
	if ex != nil {

		err = exception.NewException("Invalid int value: %s", val)
		ival = 0
	}

	return ival * rate, err

}

func ParseCharset(val string) string {
	_, ex := htmlindex.Get(val)
	if ex == nil {
		return ""

	} else {
		return val
	}
}

func StartsWith(str, end string) bool {
	return strings.HasPrefix(str, end)
}

func EndsWith(str, end string) bool {
	return strings.HasSuffix(str, end)
}

func Indent(count int) string {
	return strings.Repeat(" ", count)
}
