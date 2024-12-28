package common

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
)

/**
 * Managements I/O Multiplexing
 *  (Possible implementations include the select system call, event APIs, or threading)
 */

type Multiplexer interface {
	AddRudderState(rd rudder.Rudder, st *RudderState)

	RemoveRudderState(rd rudder.Rudder)

	GetTransporter(rd rudder.Rudder) Transporter

	ReqAccept(rd rudder.Rudder)

	ReqConnect(rd rudder.Rudder, adr net.Addr) exception.IOException

	ReqRead(rd rudder.Rudder)

	ReqWrite(rd rudder.Rudder, buf []byte, adr net.Addr, tag interface{}, lis DataConsumeListener) exception.IOException

	ReqClose(rd rudder.Rudder)

	ReqEnd(r rudder.Rudder)

	CancelRead(st *RudderState)
	CancelWrite(st *RudderState)

	NextAccept(st *RudderState)
	NextRead(st *RudderState)
	NextWrite(st *RudderState)

	Shutdown()

	IsNonBlocking() bool

	ConsumeOldestUnit(st *RudderState) bool
	CloseRudder(st *RudderState)

	IsBusy() bool
	OnBusy()
	OnFree()
}
