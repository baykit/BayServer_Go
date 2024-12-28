package multiplexer

import (
	"bayserver-core/baykit/bayserver/common"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
)

type PlainTransporter struct {
	//common.Transporter
	multiplexer    common.Multiplexer
	serverMode     bool
	traceSSL       bool
	readBufferSize int
	needHandshake  bool
	ship           ship.Ship
	closed         bool
	tmpAddress     []net.Addr
}

func NewPlainTransporter(mpx common.Multiplexer, sip ship.Ship, serverMode bool, bufsize int, traceSSL bool) *PlainTransporter {
	tp := PlainTransporter{
		multiplexer:    mpx,
		serverMode:     serverMode,
		readBufferSize: bufsize,
		traceSSL:       traceSSL,
		needHandshake:  true,
		ship:           sip,
		closed:         false,
		tmpAddress:     make([]net.Addr, 0),
	}

	var _ common.Transporter = &tp // cast check
	return &tp
}

func (tp *PlainTransporter) Init() {
}

func (tp *PlainTransporter) String() string {
	return "tp[" + tp.ship.String() + "]"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (tp *PlainTransporter) Reset() {
	tp.closed = false
}

/****************************************/
/* Implements Transporter               */
/****************************************/

func (tp *PlainTransporter) OnConnect(rd rudder.Rudder) (common.NextSocketAction, exception.IOException) {
	return tp.ship.NotifyConnect()
}

func (tp *PlainTransporter) OnRead(rd rudder.Rudder, buf []byte, addr net.Addr) (common.NextSocketAction, exception.IOException) {

	if len(buf) == 0 {
		return tp.ship.NotifyEof(), nil

	} else {
		nextAct, err := tp.ship.NotifyRead(buf)

		if err != nil {
			if _, ok := err.(exception2.UpgradeException); ok {
				baylog.Debug("%s Protocol upgrade", tp.ship)
				return tp.ship.NotifyRead(buf)

			} else if _, ok := err.(exception2.ProtocolException); ok {
				closing, ioerr := tp.ship.NotifyProtocolError(err.(exception2.ProtocolException))
				if ioerr != nil {
					return -1, ioerr

				} else if !closing && tp.serverMode {
					return common.NEXT_SOCKET_ACTION_CONTINUE, nil

				} else {
					return common.NEXT_SOCKET_ACTION_CLOSE, nil
				}

			} else if _, ok := err.(exception.IOException); ok {
				// IOException which occur in notifyRead must be distinguished from
				// IOException which occur in handshake or readNonBlock.
				tp.OnError(rd, err)
				return common.NEXT_SOCKET_ACTION_CLOSE, nil

			} else {
				return -1, err
			}
		}

		return nextAct, nil
	}
}

func (tp *PlainTransporter) OnError(rd rudder.Rudder, err exception.Exception) {
	tp.ship.NotifyError(err)
}

func (tp *PlainTransporter) OnClosed(rd rudder.Rudder) {
	tp.ship.NotifyClose()
}

func (tp *PlainTransporter) ReqConnect(rd rudder.Rudder, addr net.Addr) exception.IOException {
	return tp.multiplexer.ReqConnect(rd, addr)
}

func (tp *PlainTransporter) ReqRead(rd rudder.Rudder) {
	tp.multiplexer.ReqRead(rd)
}

func (tp *PlainTransporter) ReqWrite(rd rudder.Rudder, buf []byte, addr net.Addr, tag interface{}, listener common.DataConsumeListener) exception.IOException {
	return tp.multiplexer.ReqWrite(rd, buf, addr, tag, listener)
}

func (tp *PlainTransporter) ReqClose(rd rudder.Rudder) {
	tp.closed = true
	tp.multiplexer.ReqClose(rd)
}

func (tp *PlainTransporter) CheckTimeout(rd rudder.Rudder, durationSec int) bool {
	return tp.ship.CheckTimeout(durationSec)
}

func (tp *PlainTransporter) GetReadBufferSize() int {
	return tp.readBufferSize
}

func (tp *PlainTransporter) PrintUsage(indent int) {
}

/****************************************/
/* Public methods                       */
/****************************************/

func (tp *PlainTransporter) Secure() bool {
	return false
}
