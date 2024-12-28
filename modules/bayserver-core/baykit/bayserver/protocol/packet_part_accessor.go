package protocol

import (
	"bayserver-core/baykit/bayserver/util/baylog"
)

type PacketPartAccessor struct {
	packet Packet
	start  int
	maxLen int
	pos    int
}

func NewPacketPartAccessor(pkt Packet, start int, maxLen int) *PacketPartAccessor {
	return &PacketPartAccessor{
		packet: pkt,
		start:  start,
		maxLen: maxLen,
		pos:    0,
	}
}

func (p *PacketPartAccessor) Pos() int {
	return p.pos
}

func (p *PacketPartAccessor) PutByte(b int) {
	p.PutBytes([]byte{byte(b)}, 0, 1)
}

func (p *PacketPartAccessor) PutBytes(buf []byte, ofs int, length int) {
	if length > 0 {
		p.checkWrite(length)
		for p.start+p.pos+length > len(p.packet.Buf()) {
			p.packet.Expand()
		}
		copy(p.packet.Buf()[p.start+p.pos:p.start+p.pos+length], buf[ofs:])
		p.forward(length)
	}
}

func (p *PacketPartAccessor) PutShort(val int) {
	h := val >> 8 & 0xFF
	l := val & 0xFF
	p.PutBytes([]byte{byte(h), byte(l)}, 0, 2)
}

func (p *PacketPartAccessor) PutInt(val int) {
	b1 := val >> 24 & 0xFF
	b2 := val >> 16 & 0xFF
	b3 := val >> 8 & 0xFF
	b4 := val & 0xFF
	p.PutBytes([]byte{byte(b1), byte(b2), byte(b3), byte(b4)}, 0, 4)
}

func (p *PacketPartAccessor) PutString(s string) {
	bytes := []byte(s)
	p.PutBytes(bytes, 0, len(bytes))
}

func (p *PacketPartAccessor) GetByte() int {
	buf := make([]byte, 1)
	p.GetBytes(buf, 0, 1)
	return int(buf[0]) & 0xFF
}

func (p *PacketPartAccessor) GetBytes(buf []byte, ofs int, length int) {
	p.checkRead(length)
	copy(buf[ofs:], p.packet.Buf()[p.start+p.pos:p.start+p.pos+length])
	p.pos += length
}

func (p *PacketPartAccessor) GetShort() int {
	h := p.GetByte()
	l := p.GetByte()
	return h<<8 | l
}

func (p *PacketPartAccessor) GetInt() int {
	b1 := p.GetByte()
	b2 := p.GetByte()
	b3 := p.GetByte()
	b4 := p.GetByte()
	return b1<<24 | b2<<16 | b3<<8 | b4
}

func (p *PacketPartAccessor) checkRead(len int) {
	var max int
	if p.maxLen >= 0 {
		max = p.maxLen
	} else {
		max = p.packet.BufLen() - p.start
	}
	if p.pos+len > max {
		baylog.Fatal("Index out of range")
		panic("Index out of range")
	}
}

func (p *PacketPartAccessor) checkWrite(len int) {
	if p.maxLen > 0 && p.pos+len > p.maxLen {
		baylog.Fatal("Index out of range")
		panic("Index out of range")
	}
}

func (p *PacketPartAccessor) forward(len int) {
	p.pos += len
	if p.start+p.pos > p.packet.BufLen() {
		p.packet.SetBufLen(p.start + p.pos)
	}
}
