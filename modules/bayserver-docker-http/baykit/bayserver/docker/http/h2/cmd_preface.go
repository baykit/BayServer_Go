package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
 * Preface is dummy command and packet
 *
 *   packet is not in frame format but raw data: "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
 */

var PREFACE_BYTES = []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")

type CmdPreface struct {
	*H2CommandBase
	Protocol string
}

func NewCmdPreface(streamId int, flags *H2Flags) protocol.Command {
	c := &CmdPreface{
		H2CommandBase: NewH2CommandBase(H2_TYPE_PREFACE, flags, streamId),
	}
	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdPreface) Unpack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	prefaceData := make([]byte, 24)
	acc.GetBytes(prefaceData, 0, len(prefaceData))
	c.Protocol = string(prefaceData[6:14])
	return nil
}

func (c *CmdPreface) Pack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	acc.PutBytes(PREFACE_BYTES, 0, len(PREFACE_BYTES))
	return nil
}

func (c *CmdPreface) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandlePreface(c)
}
