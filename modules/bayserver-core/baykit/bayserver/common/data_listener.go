package common

import (
	"bayserver-core/baykit/bayserver/common/exception"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"net"
)

type DataListener interface {
	NotifyHandshakeDone(protocol string) (int, exception2.IOException)
	NotifyConnect() (int, exception2.IOException)
	NotifyRead(buf []byte, adr net.Addr) (int, exception2.IOException)
	NotifyEof() int
	NotifyError(err exception2.Exception)
	NotifyProtocolError(err exception.ProtocolException) bool
	NotifyClose()
	CheckTimeout(durationSec int) bool
}
