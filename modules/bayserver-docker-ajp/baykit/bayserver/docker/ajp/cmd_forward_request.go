package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"strconv"
)

/**
 * AJP protocol
 *    https://tomcat.apache.org/connectors-doc/ajp/ajpv13a.html
 *
 * AJP13_FORWARD_REQUEST :=
 *     prefix_code      (byte) 0x02 = JK_AJP13_FORWARD_REQUEST
 *     method           (byte)
 *     protocol         (string)
 *     req_uri          (string)
 *     remote_addr      (string)
 *     remote_host      (string)
 *     server_name      (string)
 *     server_port      (integer)
 *     is_ssl           (boolean)
 *     num_headers      (integer)
 *     request_headers *(req_header_name req_header_value)
 *     attributes      *(attribut_name attribute_value)
 *     request_terminator (byte) OxFF
 */

var methods = map[int]string{
	1:  "OPTIONS",
	2:  "GET",
	3:  "HEAD",
	4:  "POST",
	5:  "PUT",
	6:  "DELETE",
	7:  "TRACE",
	8:  "PROPFIND",
	9:  "PROPPATCH",
	10: "MKCOL",
	11: "COPY",
	12: "MOVE",
	13: "LOCK",
	14: "UNLOCK",
	15: "ACL",
	16: "REPORT",
	17: "VERSION_CONTROL",
	18: "CHECKIN",
	19: "CHECKOUT",
	20: "UNCHECKOUT",
	21: "SEARCH",
	22: "MKWORKSPACE",
	23: "UPDATE",
	24: "LABEL",
	25: "MERGE",
	26: "BASELINE_CONTROL",
	27: "MKACTIVITY",
}

func GetMethodCode(method string) int {
	for code, m := range methods {
		if m == method {
			return code
		}
	}
	return -1
}

var wellKnownHeaders = map[int]string{
	0xA001: "Accept",
	0xA002: "Accept-Charset",
	0xA003: "Accept-Encoding",
	0xA004: "Accept-Language",
	0xA005: "Authorization",
	0xA006: "Connection",
	0xA007: "Content-Type",
	0xA008: "Content-Length",
	0xA009: "Cookie",
	0xA00A: "Cookie2",
	0xA00B: "Host",
	0xA00C: "Pragma",
	0xA00D: "Referer",
	0xA00E: "User-Agent",
}

func GetWellKnownHeaderCode(name string) int {
	for code, h := range wellKnownHeaders {
		if h == name {
			return code
		}
	}
	return -1
}

var attributeNames = map[int]string{
	0x01: "?context",
	0x02: "?servlet_path",
	0x03: "?remote_user",
	0x04: "?auth_type",
	0x05: "?query_string",
	0x06: "?route",
	0x07: "?ssl_cert",
	0x08: "?ssl_cipher",
	0x09: "?ssl_session",
	0x0A: "?req_attribute",
	0x0B: "?ssl_key_size",
	0x0C: "?secret",
	0x0D: "?stored_method",
}

func GetAttributeCode(atr string) int {
	for code, a := range attributeNames {
		if a == atr {
			return code
		}
	}
	return -1
}

type CmdForwardRequest struct {
	*AjpCommandBase
	Method     string
	Protocol   string
	ReqUri     string
	RemoteAddr string
	RemoteHost string
	ServerName string
	ServerPort int
	IsSsl      bool
	Headers    *headers.Headers
	Attributes map[string]string
}

func NewCmdForwardRequest() *CmdForwardRequest {
	c := CmdForwardRequest{
		AjpCommandBase: NewAjpCommandBase(AJP_TYPE_FORWARD_REQUEST, true),
		Headers:        headers.NewHeaders(),
		Attributes:     map[string]string{},
	}
	var _ protocol.Command = &c // cast check
	var _ AjpCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdForwardRequest) Unpack(pkt protocol.Packet) exception2.IOException {
	c.AjpCommandBase.Unpack(pkt.(*AjpPacket))

	acc := pkt.(*AjpPacket).NewAjpDataAccessor()
	acc.GetByte() // prefix code
	c.Method = methods[acc.GetByte()]
	c.Protocol = acc.GetString()
	c.ReqUri = acc.GetString()
	c.RemoteAddr = acc.GetString()
	c.RemoteHost = acc.GetString()
	c.ServerName = acc.GetString()
	c.ServerPort = acc.GetShort()
	c.IsSsl = acc.GetByte() == 1

	ioerr := c.readRequestHeaders(acc)
	if ioerr != nil {
		return ioerr
	}
	ioerr = c.readAttributes(acc)
	if ioerr != nil {
		return ioerr
	}

	return nil
}

func (c *CmdForwardRequest) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.(*AjpPacket).NewAjpDataAccessor()

	acc.PutByte(c.Type())
	acc.PutByte(GetMethodCode(c.Method))
	acc.PutString(c.Protocol)
	acc.PutString(c.ReqUri)
	acc.PutString(c.RemoteAddr)
	acc.PutString(c.RemoteAddr)
	acc.PutString(c.ServerName)
	acc.PutShort(c.ServerPort)
	if c.IsSsl {
		acc.PutByte(1)
	} else {
		acc.PutByte(0)
	}

	c.writeRequestHeaders(acc)
	c.writeAttributes(acc)
	// must be called from last line
	c.AjpCommandBase.Pack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdForwardRequest) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(AjpCommandHandler).HandleForwardRequest(c)
}

func (c *CmdForwardRequest) readRequestHeaders(acc *AjpAccessor) exception2.IOException {
	count := acc.GetShort()
	for i := 0; i < count; i++ {
		code := acc.GetShort()
		var name string
		if code >= 0xa000 {
			name = wellKnownHeaders[code]
			if name == "" {
				return exception.NewProtocolException("Invalid header")
			}

		} else {
			name = acc.GetStringByLen(code)
		}

		value := acc.GetString()
		c.Headers.Add(name, value)
	}
	return nil
}

func (c *CmdForwardRequest) readAttributes(acc *AjpAccessor) exception2.IOException {

	for {
		code := acc.GetByte()
		var name string
		if code == 0xff {
			break

		} else if code == 0x0a {
			name = acc.GetString()

		} else {
			name = attributeNames[code]
			if name == "" {
				return exception.NewProtocolException("Invalid attribute: code=%d", code)
			}
		}

		if code == 0x0b { // "?ssl_key_size"
			value := acc.GetShort()
			c.Attributes[name] = strconv.Itoa(value)

		} else {
			value := acc.GetString()
			c.Attributes[name] = value
		}
	}
	return nil
}

func (c *CmdForwardRequest) writeRequestHeaders(acc *AjpAccessor) {
	hlist := make([][]string, 0)
	for _, name := range c.Headers.HeaderNames() {
		for _, value := range c.Headers.HeaderValues(name) {
			hlist = append(hlist, []string{name, value})
		}
	}

	acc.PutShort(len(hlist))
	for _, hdr := range hlist {
		code := GetWellKnownHeaderCode(hdr[0])
		if code != -1 {
			acc.PutShort(code)

		} else {
			acc.PutString(hdr[0])
		}
		acc.PutString(hdr[1])
	}
}

func (c *CmdForwardRequest) writeAttributes(acc *AjpAccessor) {
	for name, value := range c.Attributes {
		code := GetAttributeCode(name)

		if code != -1 {
			acc.PutByte(code)
		} else {
			acc.PutString(name)
		}
		acc.PutString(value)
	}
	acc.PutByte(0xFF) // termination code
}
