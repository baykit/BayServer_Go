package ajp

import "bayserver-core/baykit/bayserver/protocol"

type AjpAccessor struct {
	*protocol.PacketPartAccessor
}

func NewAjpAccessor(pkt protocol.Packet, start int, maxLen int) *AjpAccessor {
	return &AjpAccessor{
		protocol.NewPacketPartAccessor(pkt, start, maxLen),
	}
}

func (a *AjpAccessor) PutString(str string) {
	if str == "" {
		a.PutShort(0xffff)

	} else {
		a.PutShort(len(str))
		a.PacketPartAccessor.PutString(str)
		a.PutByte(0)
	}
}

func (a *AjpAccessor) GetStringByLen(length int) string {
	if length == 0xffff {
		return ""

	}

	buf := make([]byte, length)
	a.GetBytes(buf, 0, length)
	a.GetByte() // null terminator

	return string(buf)
}

func (a *AjpAccessor) GetString() string {
	return a.GetStringByLen(a.GetShort())
}
