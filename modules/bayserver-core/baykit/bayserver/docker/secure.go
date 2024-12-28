package docker

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
)

type Secure interface {
	Docker

	SetAppProtocols(protocols []string)

	ReloadCert()

	NewTransporter(agtId int, sip ship.Ship) common.Transporter

	GetSecureConn(conn net.Conn) (net.Conn, exception.IOException)
}
