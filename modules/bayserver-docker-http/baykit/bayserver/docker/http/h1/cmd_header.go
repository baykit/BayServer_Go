package h1

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/charutil"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"strconv"
	"strings"
)

/**
 * Header format
 *
 *
 *        generic-message = start-line
 *                           *(message-header CRLF)
 *                           CRLF
 *                           [ message-body ]
 *        start-line      = Request-Line | Status-Line
 *
 *
 *        message-header = field-name ":" [ field-value ]
 *        field-name     = token
 *        field-value    = *( field-content | LWS )
 *        field-content  = <the OCTETs making up the field-value
 *                         and consisting of either *TEXT or combinations
 *                         of token, separators, and quoted-string>
 */

const STATE_READ_FIRST_LINE = 1
const STATE_READ_MESSAGE_HEADERS = 2

type CmdHeader struct {
	*impl.CommandBase
	headers [][]string
	req     bool
	method  string
	uri     string
	version string
	status  int
}

func NewCmdHeader(req bool) *CmdHeader {
	c := CmdHeader{
		CommandBase: impl.NewCommandBase(H1_TYPE_HEADER),
	}
	var _ protocol.Command = &c // cast check
	var _ H1Command = &c        // cast check
	c.req = req
	return &c
}

func NewReqHeader(method string, uri string, version string) *CmdHeader {
	h := NewCmdHeader(true)
	h.method = method
	h.uri = uri
	h.version = version
	return h
}

func NewResHeader(headers *headers.Headers, version string) *CmdHeader {
	h := NewCmdHeader(false)
	h.version = version
	h.status = headers.Status()
	for _, name := range headers.HeaderNames() {
		for _, value := range headers.HeaderValues(name) {
			h.AddHeader(name, value)
		}
	}
	return h
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdHeader) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.NewDataAccessor()

	if c.req {
		c.packRequestLine(acc)
	} else {
		c.packStatusLine(acc)
	}

	for _, nv := range c.headers {
		c.packMessageHeader(acc, nv[0], nv[1])
	}

	c.packEndHeader(acc)
	return nil
}

func (c *CmdHeader) Unpack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.NewDataAccessor()
	pos := 0
	dataLen := pkt.DataLen()
	state := STATE_READ_FIRST_LINE

	lineStartPos := 0
	lineLen := 0

loop:
	for pos = 0; pos < dataLen; pos++ {
		b := acc.GetByte()
		switch b {
		case '\r':

		case '\n':
			if lineLen == 0 {
				break loop
			}

			var ioerr exception2.IOException
			if state == STATE_READ_FIRST_LINE {
				if c.req {
					ioerr = c.unpackRequestLine(pkt.Buf(), lineStartPos, lineLen)
				} else {
					ioerr = c.unpackStatusLine(pkt.Buf(), lineStartPos, lineLen)
				}
				state = STATE_READ_MESSAGE_HEADERS

			} else {
				ioerr = c.unpackMessageHeader(pkt.Buf(), lineStartPos, lineLen)
			}

			if ioerr != nil {
				return ioerr
			}
			lineLen = 0
			lineStartPos = pos + 1

		default:
			lineLen++
		}
	}

	if state == STATE_READ_FIRST_LINE {
		return exception.NewProtocolException("Invalid HTTP header format: %s", string(pkt.Buf()[0:dataLen]))
	}
	return nil
}

func (c *CmdHeader) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(H1CommandHandler).HandleHeader(c)
}

/****************************************/
/* Public functions                     */
/****************************************/

func (c *CmdHeader) AddHeader(name string, value string) {
	c.headers = append(c.headers, []string{name, value})
}

func (c *CmdHeader) SetHeader(name string, value string) {
	for _, nv := range c.headers {
		if strings.EqualFold(nv[0], name) {
			nv[1] = value
			return
		}
	}
	c.AddHeader(name, value)
}

/****************************************/
/* Private functions                    */
/****************************************/

func (c *CmdHeader) unpackRequestLine(buf []byte, start int, length int) exception2.IOException {
	line := string(buf[start : start+length])
	items := strings.Split(line, " ")
	if len(items) != 3 {
		return exception.NewProtocolException(baymessage.Get(symbol.HTP_INVALID_FIRST_LINE, line))
	}
	c.method = items[0]
	c.uri = items[1]
	c.version = items[2]
	return nil
}

func (c *CmdHeader) unpackStatusLine(buf []byte, start int, length int) exception2.IOException {
	line := string(buf[start : start+length])
	parts := strings.Split(line, " ")

	if len(parts) < 2 {
		return exception.NewProtocolException(symbol.HTP_INVALID_FIRST_LINE, line)
	}

	status := parts[1]
	var err error
	c.status, err = strconv.Atoi(status)
	if err != nil {
		return exception.NewProtocolException(symbol.HTP_INVALID_FIRST_LINE, line)
	}
	return nil
}

func (c *CmdHeader) unpackMessageHeader(bytes []byte, start int, length int) exception2.IOException {
	buf := make([]byte, length)
	readName := true
	name := ""
	value := ""
	skipping := true
	pos := 0

	for i := 0; i < length; i++ {
		var b = bytes[start+i]
		if skipping && b == ' ' {
			continue

		} else if readName && b == ':' {
			// header name completed
			name = string(buf[0:pos])
			pos = 0
			skipping = true
			readName = false

		} else {
			if readName {
				// make the case of header name be lower force
				buf[pos] = charutil.Lower(b)

			} else {
				buf[pos] = b
			}
			pos++
			skipping = false
		}
	}

	if name == "" {
		baylog.Debug("Invalid message header: %s", string(bytes[start:start+length]))
		return exception.NewProtocolException(baymessage.Get(symbol.HTP_INVALID_HEADER_FORMAT, ""))
	}

	value = string(buf[0:pos])

	c.AddHeader(name, value)
	return nil
}

func (c *CmdHeader) packRequestLine(acc *protocol.PacketPartAccessor) {
	acc.PutString(c.method)
	acc.PutBytes(SP_BYTES, 0, len(SP_BYTES))
	acc.PutString(c.uri)
	acc.PutBytes(SP_BYTES, 0, len(SP_BYTES))
	acc.PutString(c.version)
	acc.PutBytes(CRLF_BYTES, 0, len(CRLF_BYTES))
}

func (c *CmdHeader) packStatusLine(acc *protocol.PacketPartAccessor) {
	desc := httpstatus.GetDescription(c.status)

	if c.version != "" && strings.ToLower(c.version) == "http/1.1" {
		acc.PutBytes(HTTP_11_BYTES, 0, len(HTTP_11_BYTES))
	} else {
		acc.PutBytes(HTTP_10_BYTES, 0, len(HTTP_10_BYTES))
	}

	// status
	acc.PutBytes(SP_BYTES, 0, len(SP_BYTES))
	acc.PutString(strconv.Itoa(c.status))
	acc.PutBytes(SP_BYTES, 0, len(SP_BYTES))
	acc.PutString(desc)
	acc.PutBytes(CRLF_BYTES, 0, len(CRLF_BYTES))
}

func (c *CmdHeader) packMessageHeader(acc *protocol.PacketPartAccessor, name string, value string) {
	acc.PutString(name)
	acc.PutBytes(headers.HEADER_SEPARATOR_BYTES, 0, len(headers.HEADER_SEPARATOR_BYTES))
	acc.PutString(value)
	acc.PutBytes(CRLF_BYTES, 0, len(CRLF_BYTES))
}

func (c *CmdHeader) packEndHeader(acc *protocol.PacketPartAccessor) {
	acc.PutBytes(CRLF_BYTES, 0, len(CRLF_BYTES))
}
