package fcgi

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 *
 * Fast CGI response rule
 *
 *   (StdOut | StdErr)* EndRequest
 *
 */

type FcgCommandUnpacker struct {
	impl.CommandUnpackerImpl
	Handler FcgCommandHandler
}

func NewFcgCommandUnpacker(handler FcgCommandHandler) *FcgCommandUnpacker {
	cu := FcgCommandUnpacker{
		Handler: handler,
	}
	cu.Reset()
	return &cu
}

/****************************************/
/* Implements CommandUnpacker           */
/****************************************/

func (cu *FcgCommandUnpacker) PacketReceived(pkt protocol.Packet) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("fcg: read packet type=%d length=%d", pkt.Type(), pkt.DataLen())

	reqId := pkt.(*FcgPacket).ReqId
	var cmd protocol.Command
	switch pkt.Type() {
	case FCG_TYPE_BEGIN_REQUEST:
		cmd = NewCmdBeginRequest(reqId)

	case FCG_TYPE_END_REQUEST:
		cmd = NewCmdEndRequest(reqId)

	case FCG_TYPE_PARAMS:
		cmd = NewCmdParams(reqId)

	case FCG_TYPE_STDIN:
		cmd = NewCmdStdIn(reqId)

	case FCG_TYPE_STDOUT:
		cmd = NewCmdStdOut(reqId, nil, 0, 0)

	case FCG_TYPE_STDERR:
		cmd = NewCmdStdErr(reqId)

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
