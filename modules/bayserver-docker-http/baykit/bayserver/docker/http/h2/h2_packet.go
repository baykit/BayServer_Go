package h2

import (
	"bayserver-core/baykit/bayserver/protocol"
	"strconv"
)

/**
 * Http2 spec
 *   https://www.rfc-editor.org/rfc/rfc7540.txt
 *
 * Http2 Frame format
 * +-----------------------------------------------+
 * |                 Length (24)                   |
 * +---------------+---------------+---------------+
 * |   Type (8)    |   Flags (8)   |
 * +-+-+-----------+---------------+-------------------------------+
 * |R|                 Stream Identifier (31)                      |
 * +=+=============================================================+
 * |                   Frame Payload (0...)                      ...
 * +---------------------------------------------------------------+
 */

const MAX_PAYLOAD_MAXLEN = 0x00FFFFFF     // = 2^24-1 = 16777215 = 16MB-1
const DEFAULT_PAYLOAD_MAXLEN = 0x00004000 // = 2^14 = 16384 = 16KB
const FRAME_HEADER_LEN = 9

type H2Packet struct {
	protocol.PacketImpl

	Flags    *H2Flags
	StreamId int
}

func NewH2Packet(typ int) *H2Packet {
	p := H2Packet{}
	p.ConstructPacket(typ, FRAME_HEADER_LEN, DEFAULT_PAYLOAD_MAXLEN)
	return &p
}

func (p *H2Packet) String() string {
	return "H2Packet(" + strconv.Itoa(p.Type()) + ")"
}

func (p *H2Packet) PackHeader() {
	acc := p.NewH2HeaderAccessor()
	acc.PutInt24(p.DataLen())
	acc.PutByte(p.Type())
	acc.PutByte(p.Flags.Flags)
	acc.PutInt(ExtractInt31(p.StreamId))
}

func (p *H2Packet) NewH2HeaderAccessor() *H2HeaderAccessor {
	return NewH2HeaderAccessor(p, 0, p.HeaderLen())
}

func (p *H2Packet) NewH2DataAccessor() *H2DataAccessor {
	return NewH2DataAccessor(p, p.HeaderLen(), -1)
}

/****************************************/
/* Static functions                     */
/****************************************/

func ExtractInt31(val int) int {
	return val & 0x7FFFFFFF
}

func ExtractFlag(val int) int {
	return ((val & 0x80000000) >> 31) & 1
}

func ConsolidateFlagAndInt32(flag int, val int) int {
	return (flag&1)<<31 | (val & 0x7FFFFFFF)
}

func MakeStreamDependency32(excluded bool, dep int) int {
	val := 0
	if excluded {
		val = 1
	}
	return val<<31 | ExtractInt31(dep)
}
