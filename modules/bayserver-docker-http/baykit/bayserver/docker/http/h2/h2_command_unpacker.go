package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type H2CommandUnPacker struct {
	impl.CommandUnpackerImpl
	cmdHandler H2CommandHandler
}

func NewH2CommandUnpacker(handler H2CommandHandler) *H2CommandUnPacker {
	cu := H2CommandUnPacker{}
	cu.cmdHandler = handler
	cu.Reset()
	return &cu
}

/****************************************/
/* Implements CommandUnpacker           */
/****************************************/

func (cu *H2CommandUnPacker) PacketReceived(pkt protocol.Packet) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("h1: read packet type=%d length=%d", pkt.Type(), pkt.DataLen())
	h2pkt := pkt.(*H2Packet)

	var cmd protocol.Command
	switch pkt.Type() {
	case H2_TYPE_PREFACE:
		cmd = NewCmdPreface(h2pkt.StreamId, h2pkt.Flags)

	case H2_TYPE_HEADERS:
		cmd = NewCmdHeaders(h2pkt.StreamId, h2pkt.Flags)

	case H2_TYPE_PRIORITY:
		cmd = NewCmdPriority(h2pkt.StreamId, h2pkt.Flags)

	case H2_TYPE_SETTINGS:
		cmd = NewCmdSettings(h2pkt.StreamId, h2pkt.Flags)

	case H2_TYPE_WINDOW_UPDATE:
		cmd = NewCmdWindowUpdate(h2pkt.StreamId, h2pkt.Flags)

	case H2_TYPE_DATA:
		cmd = NewCmdData(h2pkt.StreamId, h2pkt.Flags, nil, 0, 0)

	case H2_TYPE_GOAWAY:
		cmd = NewCmdGoAway(h2pkt.StreamId, h2pkt.Flags)

	case H2_TYPE_PING:
		cmd = NewCmdPing(h2pkt.StreamId, h2pkt.Flags, nil)

	case H2_TYPE_RST_STREAM:
		cmd = NewCmdRstStream(h2pkt.StreamId, h2pkt.Flags)

	default:
		baylog.FatalE(exception2.NewSink("Illegal State"), "")
		panic("Error")
	}

	ioerr := cmd.Unpack(pkt)
	if ioerr != nil {
		return -1, ioerr
	}

	return cmd.Handle(cu.cmdHandler)

}
