package util

import "strings"

const MATCH_TYPE_ALL = 1
const MATCH_TYPE_EXACT = 2
const MATCH_TYPE_DOMAIN = 3

type HostMatcher struct {
	matchType int
	host      string
	domain    string
}

func NewHostMatcher(host string) *HostMatcher {
	hm := HostMatcher{}
	if host == "*" {
		hm.matchType = MATCH_TYPE_ALL

	} else if strings.HasPrefix(host, "*.") {
		hm.matchType = MATCH_TYPE_DOMAIN
		hm.domain = host[2:]

	} else {
		hm.matchType = MATCH_TYPE_EXACT
		hm.host = host
	}
	return &hm
}

func (m *HostMatcher) Match(remoteHost string) bool {
	if m.matchType == MATCH_TYPE_ALL {
		// all match
		return true
	}

	if remoteHost == "" {
		return false
	}

	if m.matchType == MATCH_TYPE_EXACT {
		// exact match
		return remoteHost == m.host

	} else {
		// domain match
		return strings.HasSuffix(remoteHost, m.domain)
	}
}
