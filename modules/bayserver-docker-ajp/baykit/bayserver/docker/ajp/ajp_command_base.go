package ajp

import (
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
)

type AjpCommandBase struct {
	*impl.CommandBase
	toServer bool
}

func NewAjpCommandBase(typ int, toServer bool) *AjpCommandBase {
	h := &AjpCommandBase{
		CommandBase: impl.NewCommandBase(typ),
		toServer:    toServer,
	}

	var _ AjpCommand = h // implement check
	return h
}

func (h *AjpCommandBase) ToServer() bool {
	return h.toServer
}

func (h *AjpCommandBase) SetToServer(toServer bool) {
	h.toServer = toServer
}

func (h *AjpCommandBase) Unpack(pkt *AjpPacket) {
	if pkt.Type() != h.Type() {
		baylog.Fatal("IllegalArgument")
	}
	h.toServer = pkt.toServer
}

/**
 * Base class method must be called from last line of override method since header cannot be packed before data is constructed
 */

func (h *AjpCommandBase) Pack(pkt *AjpPacket) {
	if pkt.Type() != h.Type() {
		baylog.Fatal("IllegalArgument")
	}
	pkt.toServer = h.toServer
	h.packHeader(pkt)
}

func (h *AjpCommandBase) packHeader(pkt *AjpPacket) {
	acc := pkt.NewAjpHeaderAccessor()
	if pkt.toServer {
		acc.PutByte(0x12)
		acc.PutByte(0x34)

	} else {
		acc.PutByte('A')
		acc.PutByte('B')
	}
	acc.PutByte((pkt.DataLen() >> 8) & 0xff)
	acc.PutByte(pkt.DataLen() & 0xff)
}
