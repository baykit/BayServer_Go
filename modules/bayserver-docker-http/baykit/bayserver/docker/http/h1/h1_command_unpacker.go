package h1

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type H1CommandUnpacker struct {
	impl.CommandUnpackerImpl
	ServerMode bool
	Handler    H1CommandHandler
}

func NewH1CommandUnpacker(handler H1CommandHandler, serverMode bool) *H1CommandUnpacker {
	cu := H1CommandUnpacker{}
	cu.Handler = handler
	cu.ServerMode = serverMode
	cu.Reset()
	return &cu
}

/****************************************/
/* Implements CommandUnpacker           */
/****************************************/

func (cu *H1CommandUnpacker) PacketReceived(pkt protocol.Packet) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("h1: read packet type=%d length=%d", pkt.Type(), pkt.DataLen())

	var cmd protocol.Command
	switch pkt.Type() {
	case H1_TYPE_HEADER:
		cmd = NewCmdHeader(cu.ServerMode)

	case H1_TYPE_CONTENT:
		cmd = NewCmdContent(nil, 0, 0)

	default:
		baylog.FatalE(exception2.NewSink("Illegal State"), "")
		panic("Error")
	}

	ioerr := cmd.Unpack(pkt)
	if ioerr != nil {
		return -1, ioerr
	}

	return cmd.Handle(cu.Handler)

}

func (cu *H1CommandUnpacker) ReqFinished() bool {
	return cu.Handler.ReqFinished()
}
