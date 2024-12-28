package protocol

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/util/exception"
	"fmt"
)

type Command interface {
	fmt.Stringer
	Type() int
	Unpack(pkt Packet) exception.IOException
	Pack(pkt Packet) exception.IOException
	Handle(h CommandHandler) (common.NextSocketAction, exception.IOException)
}
