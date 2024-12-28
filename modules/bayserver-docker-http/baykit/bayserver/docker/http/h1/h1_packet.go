package h1

import (
	"bayserver-core/baykit/bayserver/protocol"
	"strconv"
)

const MAX_HEADER_LEN = 0 // H1 packet does not have packet header
const MAX_DATA_LEN = 65536

/** space */
var SP_BYTES = []byte(" ")

/** Line separator */
var CRLF_BYTES = []byte("\r\n")

var HTTP_11_BYTES = []byte("HTTP/1.1")
var HTTP_10_BYTES = []byte("HTTP/1.0")

type H1Packet struct {
	protocol.PacketImpl
}

func NewH1Packet(typ int) *H1Packet {
	p := H1Packet{}
	p.ConstructPacket(typ, MAX_HEADER_LEN, MAX_DATA_LEN)
	return &p
}

func (p *H1Packet) String() string {
	return "H1Packet(" + strconv.Itoa(p.Type()) + ")"
}
