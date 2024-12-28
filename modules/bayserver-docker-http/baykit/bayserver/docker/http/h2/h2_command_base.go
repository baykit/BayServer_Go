package h2

import "bayserver-core/baykit/bayserver/protocol/impl"

type H2CommandBase struct {
	*impl.CommandBase
	flags    *H2Flags
	streamId int
}

func NewH2CommandBase(typ int, flags *H2Flags, stmId int) *H2CommandBase {
	if flags == nil {
		flags = NewH2Flags(FLAGS_NONE)
	}
	h := &H2CommandBase{
		CommandBase: impl.NewCommandBase(typ),
		flags:       flags,
		streamId:    stmId,
	}

	var _ H2Command = h // implementation check
	return h
}

func (h *H2CommandBase) Flags() *H2Flags {
	return h.flags
}

func (h *H2CommandBase) StreamId() int {
	return h.streamId
}

func (h *H2CommandBase) Unpack(pkt *H2Packet) {
	h.streamId = pkt.StreamId
	h.flags = pkt.Flags
}

func (h *H2CommandBase) Pack(pkt *H2Packet) {
	pkt.StreamId = h.streamId
	pkt.Flags = h.flags
	pkt.PackHeader()
}
