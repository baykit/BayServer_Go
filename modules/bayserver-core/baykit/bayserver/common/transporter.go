package common

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
)

/****************************************/
/* Interface Transporter                */
/****************************************/

type Transporter interface {
	Init()

	OnConnect(rd rudder.Rudder) (NextSocketAction, exception.IOException)
	OnRead(rd rudder.Rudder, buf []byte, addr net.Addr) (NextSocketAction, exception.IOException)
	OnError(rd rudder.Rudder, err exception.Exception)
	OnClosed(rd rudder.Rudder)

	ReqConnect(rd rudder.Rudder, addr net.Addr) exception.IOException
	ReqRead(rd rudder.Rudder)
	ReqWrite(rd rudder.Rudder, buf []byte, addr net.Addr, tag interface{}, listener DataConsumeListener) exception.IOException
	ReqClose(rd rudder.Rudder)

	CheckTimeout(rd rudder.Rudder, durationSec int) bool
	GetReadBufferSize() int

	PrintUsage(indent int)
}
