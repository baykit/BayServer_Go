package h1

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/util/exception"
)

type CmdEndContent struct {
	*impl.CommandBase
}

func NewCmdEndContent() protocol.Command {
	c := CmdEndContent{
		CommandBase: impl.NewCommandBase(H1_TYPE_END_CONTENT),
	}
	var _ protocol.Command = &c // cast check
	var _ H1Command = &c        // cast check

	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdEndContent) Unpack(pkt protocol.Packet) exception.IOException {
	return nil
}

func (c *CmdEndContent) Pack(pkt protocol.Packet) exception.IOException {
	return nil
}

func (c *CmdEndContent) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H1CommandHandler).HandleEndContent(c)
}
