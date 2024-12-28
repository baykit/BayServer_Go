package fcgi

import "bayserver-core/baykit/bayserver/protocol/impl"

type FcgCommandBase struct {
	*impl.CommandBase
	reqId int
}

func NewFcgCommandBase(typ int, reqId int) *FcgCommandBase {
	h := &FcgCommandBase{
		CommandBase: impl.NewCommandBase(typ),
		reqId:       reqId,
	}
	var _ FcgCommand = h // implement check
	return h
}

func (h *FcgCommandBase) ReqId() int {
	return h.reqId
}

func (h *FcgCommandBase) Unpack(pkt *FcgPacket) {
	h.reqId = pkt.ReqId
}

/**
 * Base class method must be called from last line of override method since header cannot be packed before data is constructed
 */

func (h *FcgCommandBase) Pack(pkt *FcgPacket) {
	pkt.ReqId = h.reqId
	h.packHeader(pkt)
}

func (h *FcgCommandBase) packHeader(pkt *FcgPacket) {
	acc := pkt.NewHeaderAccessor()
	acc.PutByte(pkt.Version)
	acc.PutByte(pkt.Type())
	acc.PutShort(pkt.ReqId)
	acc.PutShort(pkt.DataLen())
	acc.PutByte(0) // paddinglen
	acc.PutByte(0) // reserved
}
