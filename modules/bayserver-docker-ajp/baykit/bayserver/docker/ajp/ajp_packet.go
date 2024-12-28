package ajp

import (
	"bayserver-core/baykit/bayserver/protocol"
	"strconv"
)

/**
 * AJP Protocol
 * https://tomcat.apache.org/connectors-doc/ajp/ajpv13a.html
 *
 * AJP packet spec
 *
 *   packet:  preamble, length, body
 *   preamble:
 *        0x12, 0x34  (client->server)
 *     | 'A', 'B'     (server->client)
 *   length:
 *      2 byte
 *   body:
 *      $length byte
 *
 *
 *  Body format
 *    client->server
 *    Code     Type of Packet    Meaning
 *       2     Forward Request   Begin the request-processing cycle with the following data
 *       7     Shutdown          The web server asks the container to shut itself down.
 *       8     Ping              The web server asks the container to take control (secure login phase).
 *      10     CPing             The web server asks the container to respond quickly with a CPong.
 *    none     Data              Size (2 bytes) and corresponding body data.
 *
 *    server->client
 *    Code     Type of Packet    Meaning
 *       3     Send Body Chunk   Send a chunk of the body from the servlet container to the web server (and presumably, onto the browser).
 *       4     Send Headers      Send the response headers from the servlet container to the web server (and presumably, onto the browser).
 *       5     End Response      Marks the end of the response (and thus the request-handling cycle).
 *       6     Get Body Chunk    Get further data from the request if it hasn't all been transferred yet.
 *       9     CPong Reply       The reply to a CPing request
 *
 */

const AJP_PREAMBLE_SIZE = 4
const AJP_MAX_DATA_LEN = 8192 - AJP_PREAMBLE_SIZE
const AJP_MIN_BUF_SIZE = 1024

type AjpPacket struct {
	protocol.PacketImpl
	toServer bool
}

func NewAjpPacket(typ int) *AjpPacket {
	p := AjpPacket{}
	p.ConstructPacket(typ, AJP_PREAMBLE_SIZE, AJP_MAX_DATA_LEN)
	return &p
}

func (p *AjpPacket) Reset() {
	p.toServer = false
	p.PacketImpl.Reset()
}

func (p *AjpPacket) String() string {
	return "AjpPacket(" + strconv.Itoa(p.Type()) + ")"
}

func (p *AjpPacket) NewAjpHeaderAccessor() *AjpAccessor {
	return NewAjpAccessor(p, 0, AJP_PREAMBLE_SIZE)
}

func (p *AjpPacket) NewAjpDataAccessor() *AjpAccessor {
	return NewAjpAccessor(p, AJP_PREAMBLE_SIZE, -1)
}
