package rudder

import (
	"bayserver-core/baykit/bayserver/util/exception"
)

type ConnRudder interface {
	GetRemotePort() int
	GetRemoteAddress() string
	GetLocalAddress() string
	GetSocketReceiveBufferSize() (int, exception.IOException)
}
