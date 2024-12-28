package fcgi

import (
	"bayserver-core/baykit/bayserver/protocol"
	"strconv"
)

/**
 * FCGI spec
 *   http://www.mit.edu/~yandros/doc/specs/fcgi-spec.html
 *
 * FCGI Packet (Record) format
 *         typedef struct {
 *             unsigned char version;
 *             unsigned char type;
 *             unsigned char requestIdB1;
 *             unsigned char requestIdB0;
 *             unsigned char contentLengthB1;
 *             unsigned char contentLengthB0;
 *             unsigned char paddingLength;
 *             unsigned char reserved;
 *             unsigned char contentData[contentLength];
 *             unsigned char paddingData[paddingLength];
 *         } FCGI_Record;
 */

const FCG_PREAMBLE_SIZE = 8

const FCG_VERSION = 1
const FCG_PACKET_MAXLEN = 65535

const FCGI_NULL_REQUEST_ID = 0

type FcgPacket struct {
	protocol.PacketImpl
	Version int
	ReqId   int
}

func NewFcgPacket(typ int) *FcgPacket {
	p := FcgPacket{}
	p.ConstructPacket(typ, FCG_PREAMBLE_SIZE, FCG_PACKET_MAXLEN)
	return &p
}

func (p *FcgPacket) Reset() {
	p.Version = FCG_VERSION
	p.ReqId = 0
	p.PacketImpl.Reset()
}

func (p *FcgPacket) String() string {
	return "FcgPacket(" + strconv.Itoa(p.Type()) + ")"
}
