package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"strings"
)

/**
 * Send headers format
 *
 * AJP13_SEND_HEADERS :=
 *   prefix_code       4
 *   http_status_code  (integer)
 *   http_status_msg   (string)
 *   num_headers       (integer)
 *   response_headers *(res_header_name header_value)
 *
 * res_header_name :=
 *     sc_res_header_name | (string)   [see below for how this is parsed]
 *
 * sc_res_header_name := 0xA0 (byte)
 *
 * header_value := (string)
 */

var wellKnownResponseHeaders = map[string]int{
	"content-type":     0xA001,
	"content-language": 0xA002,
	"content-length":   0xA003,
	"date":             0xA004,
	"last-modified":    0xA005,
	"location":         0xA006,
	"set-cookie":       0xA007,
	"set-cookie2":      0xA008,
	"servlet-engine":   0xA009,
	"status":           0xA00A,
	"www-authenticate": 0xA00B,
}

func GetWellKnownResponseHeaderName(code int) string {
	for name, c := range wellKnownResponseHeaders {
		if c == code {
			return name
		}
	}
	return ""
}

type CmdSendHeaders struct {
	*AjpCommandBase
	Headers map[string][]string
	Status  int
	Desc    string
}

func NewCmdSendHeaders() *CmdSendHeaders {
	c := CmdSendHeaders{
		AjpCommandBase: NewAjpCommandBase(AJP_TYPE_SEND_HEADERS, true),
		Headers:        map[string][]string{},
		Status:         httpstatus.OK,
		Desc:           "",
	}
	var _ protocol.Command = &c // cast check
	var _ AjpCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdSendHeaders) Unpack(pkt protocol.Packet) exception2.IOException {
	c.AjpCommandBase.Unpack(pkt.(*AjpPacket))

	acc := pkt.(*AjpPacket).NewAjpDataAccessor()

	prefixCode := acc.GetByte()
	if prefixCode != AJP_TYPE_SEND_HEADERS {
		return exception.NewProtocolException("Expected SEND_HEADERS")
	}
	c.SetStatus(acc.GetShort())
	c.SetDesc(acc.GetString())

	count := acc.GetShort()
	for i := 0; i < count; i++ {
		code := acc.GetShort()
		name := GetWellKnownResponseHeaderName(code)
		if name == "" {
			// code is length
			name = acc.GetStringByLen(code)
		}
		value := acc.GetString()
		c.AddHeader(name, value)
	}
	return nil
}

func (c *CmdSendHeaders) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.(*AjpPacket).NewAjpDataAccessor()

	acc.PutByte(c.Type())
	acc.PutShort(c.Status)
	acc.PutString(httpstatus.GetDescription(c.Status))

	count := 0
	for _, values := range c.Headers {
		count += len(values)
	}

	acc.PutShort(count)
	for name, values := range c.Headers {
		code := wellKnownResponseHeaders[strings.ToLower(name)]
		for _, value := range values {
			if code > 0 {
				acc.PutShort(code)
			} else {
				acc.PutString(name)
			}
			acc.PutString(value)
		}
	}

	// must be called from last line
	c.AjpCommandBase.Pack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdSendHeaders) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(AjpCommandHandler).HandleSendHeaders(c)
}

func (c *CmdSendHeaders) SetStatus(status int) {
	c.Status = status
}

func (c *CmdSendHeaders) SetDesc(desc string) {
	c.Desc = desc
}

func (c *CmdSendHeaders) GetHeader(name string) string {

	values := c.Headers[strings.ToLower(name)]
	if values == nil || len(values) == 0 {
		return ""
	} else {
		return values[0]
	}
}

func (c *CmdSendHeaders) AddHeader(name string, value string) {
	name = strings.ToLower(name)
	values := c.Headers[name]
	if values == nil {
		values = make([]string, 0)
	}
	c.Headers[name] = append(values, value)
}
