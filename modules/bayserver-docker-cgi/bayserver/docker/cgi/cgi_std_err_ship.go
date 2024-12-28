package cgi

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/rudder"
	ship2 "bayserver-core/baykit/bayserver/ship/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"strconv"
)

type CgiStdErrShip struct {
	ship2.ShipImpl
	handler *CgiReqContentHandler
}

func NewCgiStdErrShip() *CgiStdErrShip {
	res := CgiStdErrShip{}
	res.ShipImpl.Construct()
	return &res
}

func (sip *CgiStdErrShip) Init(rd rudder.Rudder, agentId int, handler *CgiReqContentHandler) {
	sip.ShipImpl.Init(agentId, rd, nil)
	sip.handler = handler
}

func (sip *CgiStdErrShip) String() string {
	return "agt#" + strconv.Itoa(sip.AgentId()) + " err_ship#" + strconv.Itoa(sip.ShipId()) + "/" + strconv.Itoa(sip.ObjectId())
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (sip *CgiStdErrShip) Reset() {
	sip.ShipImpl.Reset()
	sip.handler = nil
}

/****************************************/
/* Implements Ship                      */
/****************************************/

func (sip *CgiStdErrShip) NotifyHandshakeDone(protocol string) (common.NextSocketAction, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return 0, nil
}

func (sip *CgiStdErrShip) NotifyConnect() (common.NextSocketAction, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return 0, nil
}

func (sip *CgiStdErrShip) NotifyRead(buf []byte) (common.NextSocketAction, exception.IOException) {
	baylog.Debug("%s CGI StdErr: read %d bytes", sip, len(buf))
	msg := string(buf)
	if len(msg) > 0 {
		baylog.Error("CGI Stderr: %s", msg)
	}

	sip.handler.Access()
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (sip *CgiStdErrShip) NotifyEof() common.NextSocketAction {
	baylog.Debug("%s EOF", sip)
	return common.NEXT_SOCKET_ACTION_CLOSE
}

func (sip *CgiStdErrShip) NotifyError(e exception.Exception) {
	baylog.DebugE(e, "")
}

func (sip *CgiStdErrShip) NotifyProtocolError(e exception2.ProtocolException) (bool, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return false, nil
}

func (sip *CgiStdErrShip) NotifyClose() {
	sip.handler.StdErrClosed()
}

func (sip *CgiStdErrShip) CheckTimeout(durationSec int) bool {
	return sip.handler.Timeout()
}
