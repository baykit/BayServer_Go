package h2

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-docker-http/baykit/bayserver/docker/http/h2/huffman"
)

type H2DataAccessor struct {
	*protocol.PacketPartAccessor
}

func NewH2DataAccessor(pkt protocol.Packet, start int, maxLen int) *H2DataAccessor {
	return &H2DataAccessor{
		protocol.NewPacketPartAccessor(pkt, start, maxLen),
	}
}

func (a *H2DataAccessor) GetHPackInt(prefix int, head *int) int {
	maxVal := 0xFF >> (8 - prefix)

	firstByte := a.GetByte()
	firstVal := firstByte & maxVal
	*head = firstByte >> prefix
	if firstVal != maxVal {
		return firstVal

	} else {
		return maxVal + a.GetHPackIntRest()
	}
}

func (a *H2DataAccessor) GetHPackIntRest() int {
	rest := 0
	for i := 0; ; i++ {
		data := a.GetByte()
		cont := (data & 0x80) != 0
		value := data & 0x7F
		rest = rest + (value << (i * 7))
		if !cont {
			break
		}
	}
	return rest
}

func (a *H2DataAccessor) GetHPackString() string {
	isHuffman := 0
	length := a.GetHPackInt(7, &isHuffman)
	data := make([]byte, length)
	a.GetBytes(data, 0, length)
	if isHuffman == 1 {
		// Huffman
		return huffman.HTreeDecode(data)

	} else {
		// ASCII
		return string(data)
	}
}

func (a *H2DataAccessor) PutHPackInt(val int, prefix int, head int) {
	maxVal := 0xFF >> (8 - prefix)
	headVal := (head << prefix) & 0xFF
	if val < maxVal {
		a.PutByte(val | headVal)

	} else {
		a.PutByte(headVal | maxVal)
		a.PutHPackIntRest(val - maxVal)
	}
}

func (a *H2DataAccessor) PutHPackIntRest(val int) {
	for {
		data := val & 0x7F
		nextVal := val >> 7
		if nextVal == 0 {
			// data end
			a.PutByte(data)
			break

		} else {
			// data continues
			a.PutByte(data | 0x80)
			val = nextVal
		}
	}
}

func (a *H2DataAccessor) PutHPackString(val string, haffman bool) {
	if haffman {
		bayserver.FatalError(exception.NewSink("Illegal Argument"))

	} else {
		a.PutHPackInt(len(val), 7, 0)
		a.PutBytes([]byte(val), 0, len(val))
	}
}
