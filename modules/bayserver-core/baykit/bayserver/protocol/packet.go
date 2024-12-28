package protocol

import (
	"bayserver-core/baykit/bayserver/util"
)

const INITIAL_BUF_SIZE = 8192 * 4

/**
 * Packet format
 *   +---------------------------+
 *   +  Header(type, length etc) +
 *   +---------------------------+
 *   +  Data(payload data)       +
 *   +---------------------------+
 */

type Packet interface {
	util.Reusable
	Type() int
	Buf() []byte
	BufLen() int
	SetBufLen(len int)
	HeaderLen() int
	MaxDataLen() int
	DataLen() int
	NewHeaderAccessor() *PacketPartAccessor
	NewDataAccessor() *PacketPartAccessor
	Expand()
}

type PacketImpl struct {
	typ        int
	buf        []byte
	bufLen     int
	headerLen  int
	maxDataLen int
}

func (p *PacketImpl) ConstructPacket(typ int, headerLen int, maxDataLen int) {
	p.typ = typ
	p.headerLen = headerLen
	p.maxDataLen = maxDataLen
	p.buf = make([]byte, INITIAL_BUF_SIZE)
	p.bufLen = headerLen
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (p *PacketImpl) Reset() {
	for i := 0; i < p.headerLen+p.DataLen(); i++ {
		p.buf[i] = 0
	}
	p.bufLen = p.headerLen
}

/****************************************/
/* Implements Packet                    */
/****************************************/

func (p *PacketImpl) Type() int {
	return p.typ
}

func (p *PacketImpl) Buf() []byte {
	return p.buf
}

func (p *PacketImpl) BufLen() int {
	return p.bufLen
}

func (p *PacketImpl) SetBufLen(len int) {
	p.bufLen = len
}

func (p *PacketImpl) HeaderLen() int {
	return p.headerLen
}

func (p *PacketImpl) DataLen() int {
	return p.bufLen - p.headerLen
}

func (p *PacketImpl) Expand() {
	p.buf = make([]byte, len(p.buf)*2)
}

func (p *PacketImpl) MaxDataLen() int {
	return p.maxDataLen
}

func (p *PacketImpl) NewHeaderAccessor() *PacketPartAccessor {
	return NewPacketPartAccessor(p, 0, p.headerLen)
}

func (p *PacketImpl) NewDataAccessor() *PacketPartAccessor {
	return NewPacketPartAccessor(p, p.headerLen, -1)
}
