package docker

import (
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
)

type Port interface {
	Docker

	Protocol() string

	Host() string

	PortNo() int

	SocketPath() string

	Address() (string, exception.IOException)

	Anchored() bool

	Secure() bool

	TimeoutSec() int

	AdditionalHeaders() [][]string

	Cities() []City

	FindCity(name string) City

	GetSecureConn(conn net.Conn) (net.Conn, exception.IOException)

	OnConnected(agentId int, rd rudder.Rudder) exception2.HttpException

	ReturnProtocolHandler(agentId int, protoHnd protocol.ProtocolHandler)

	ReturnShip(sip ship.Ship)
}
