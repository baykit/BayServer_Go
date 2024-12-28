package mimes

import (
	"strings"
)

var mimesMap = map[string]string{}

func Get(name string) string {
	return mimesMap[strings.ToLower(name)]
}

func Init(m map[string]string) {
	mimesMap = m
}
