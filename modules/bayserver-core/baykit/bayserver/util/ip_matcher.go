package util

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
)

type IpMatcher struct {
	matchAll bool
	netAdr   *net.IPNet
}

func NewIpMatcher(cidr string) (*IpMatcher, exception.IOException) {
	m := IpMatcher{}
	var ioerr exception.IOException = nil
	if cidr == "*" {
		m.matchAll = true

	} else {
		m.matchAll = false
		ioerr = m.parseIp(cidr)
	}

	return &m, ioerr
}

func (m *IpMatcher) Match(adr net.IP) bool {
	if m.matchAll {
		return true
	}

	return m.netAdr.Contains(adr)
}

func (m *IpMatcher) parseIp(cidr string) exception.IOException {

	_, netAdr, err := net.ParseCIDR(cidr)
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	}

	m.netAdr = netAdr
	return nil
}
