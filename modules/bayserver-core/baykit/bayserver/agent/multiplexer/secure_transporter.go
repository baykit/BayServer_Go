package multiplexer

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/ship"
)

type SecureTransporter struct {
	*PlainTransporter
}

func NewSecureTransporter(mpx common.Multiplexer, sip ship.Ship, serverMode bool, bufsize int, traceSSL bool) common.Transporter {
	tp := SecureTransporter{
		PlainTransporter: NewPlainTransporter(mpx, sip, serverMode, bufsize, traceSSL),
	}
	return &tp
}

func (tp *SecureTransporter) String() string {
	return "stp[" + tp.ship.String() + "]"
}

/****************************************/
/* Implements Transporter               */
/****************************************/
