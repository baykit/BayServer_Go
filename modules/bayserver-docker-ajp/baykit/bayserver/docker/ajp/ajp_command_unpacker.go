package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type AjpCommandUnpacker struct {
	impl.CommandUnpackerImpl
	Handler AjpCommandHandler
}

func NewAjpCommandUnpacker(handler AjpCommandHandler) *AjpCommandUnpacker {
	cu := AjpCommandUnpacker{
		Handler: handler,
	}
	cu.Reset()
	return &cu
}

/****************************************/
/* Implements CommandUnpacker           */
/****************************************/

func (cu *AjpCommandUnpacker) PacketReceived(pkt protocol.Packet) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("ajp read packet type=%d length=%d", pkt.Type(), pkt.DataLen())

	var cmd protocol.Command
	switch pkt.Type() {
	case AJP_TYPE_DATA:
		cmd = NewCmdData(nil, 0, 0)

	case AJP_TYPE_FORWARD_REQUEST:
		cmd = NewCmdForwardRequest()

	case AJP_TYPE_SEND_BODY_CHUNK:
		cmd = NewCmdSendBodyChunk(nil, 0, 0)

	case AJP_TYPE_SEND_HEADERS:
		cmd = NewCmdSendHeaders()

	case AJP_TYPE_END_RESPONSE:
		cmd = NewCmdEndResponse()

	case AJP_TYPE_SHUTDOWN:
		cmd = NewCmdShutdown()

	case AJP_TYPE_GET_BODY_CHUNK:
		cmd = NewCmdGetBodyChunk()

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

func (cu *AjpCommandUnpacker) NeedData() bool {
	return cu.Handler.NeedData()
}
