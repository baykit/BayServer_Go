package fcgi

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * FCGI spec
 *   http://www.mit.edu/~yandros/doc/specs/fcgi-spec.html
 *
 *
 * Params command format (Name-Value list)
 *
 *         typedef struct {
 *             unsigned char nameLengthB0;  // nameLengthB0  >> 7 == 0
 *             unsigned char valueLengthB0; // valueLengthB0 >> 7 == 0
 *             unsigned char nameData[nameLength];
 *             unsigned char valueData[valueLength];
 *         } FCGI_NameValuePair11;
 *
 *         typedef struct {
 *             unsigned char nameLengthB0;  // nameLengthB0  >> 7 == 0
 *             unsigned char valueLengthB3; // valueLengthB3 >> 7 == 1
 *             unsigned char valueLengthB2;
 *             unsigned char valueLengthB1;
 *             unsigned char valueLengthB0;
 *             unsigned char nameData[nameLength];
 *             unsigned char valueData[valueLength
 *                     ((B3 & 0x7f) << 24) + (B2 << 16) + (B1 << 8) + B0];
 *         } FCGI_NameValuePair14;
 *
 *         typedef struct {
 *             unsigned char nameLengthB3;  // nameLengthB3  >> 7 == 1
 *             unsigned char nameLengthB2;
 *             unsigned char nameLengthB1;
 *             unsigned char nameLengthB0;
 *             unsigned char valueLengthB0; // valueLengthB0 >> 7 == 0
 *             unsigned char nameData[nameLength
 *                     ((B3 & 0x7f) << 24) + (B2 << 16) + (B1 << 8) + B0];
 *             unsigned char valueData[valueLength];
 *         } FCGI_NameValuePair41;
 *
 *         typedef struct {
 *             unsigned char nameLengthB3;  // nameLengthB3  >> 7 == 1
 *             unsigned char nameLengthB2;
 *             unsigned char nameLengthB1;
 *             unsigned char nameLengthB0;
 *             unsigned char valueLengthB3; // valueLengthB3 >> 7 == 1
 *             unsigned char valueLengthB2;
 *             unsigned char valueLengthB1;
 *             unsigned char valueLengthB0;
 *             unsigned char nameData[nameLength
 *                     ((B3 & 0x7f) << 24) + (B2 << 16) + (B1 << 8) + B0];
 *             unsigned char valueData[valueLength
 *                     ((B3 & 0x7f) << 24) + (B2 << 16) + (B1 << 8) + B0];
 *         } FCGI_NameValuePair44;
 *
 */

type CmdParams struct {
	*FcgCommandBase
	Params [][]string
}

func NewCmdParams(reqId int) *CmdParams {
	c := CmdParams{
		FcgCommandBase: NewFcgCommandBase(FCG_TYPE_PARAMS, reqId),
		Params:         make([][]string, 0),
	}
	var _ protocol.Command = &c // cast check
	var _ FcgCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdParams) Unpack(pkt protocol.Packet) exception2.IOException {
	c.FcgCommandBase.Unpack(pkt.(*FcgPacket))

	acc := pkt.NewDataAccessor()
	for acc.Pos() < pkt.DataLen() {
		nameLen := c.readLength(acc)
		valueLen := c.readLength(acc)
		data := make([]byte, nameLen)
		acc.GetBytes(data, 0, len(data))
		name := string(data)

		data = make([]byte, valueLen)
		acc.GetBytes(data, 0, len(data))
		value := string(data)

		c.addParam(name, value)
	}
	return nil
}

func (c *CmdParams) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.NewDataAccessor()

	for _, nv := range c.Params {
		name := []byte(nv[0])
		value := []byte(nv[1])
		nameLen := len(name)
		valueLen := len(value)

		c.writeLength(nameLen, acc)
		c.writeLength(valueLen, acc)

		acc.PutBytes(name, 0, nameLen)
		acc.PutBytes(value, 0, valueLen)
	}

	// must be called from last line
	c.FcgCommandBase.Pack(pkt.(*FcgPacket))
	return nil
}

func (c *CmdParams) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(FcgCommandHandler).HandleParams(c)
}

/****************************************/
/* Private functions                    */
/****************************************/

func (c *CmdParams) readLength(acc *protocol.PacketPartAccessor) int {
	len1 := acc.GetByte()
	if len1>>7 == 0 {
		return len1

	} else {
		len2 := acc.GetByte()
		len3 := acc.GetByte()
		len4 := acc.GetByte()
		return ((len1 & 0x7f) << 24) | (len2 << 16) | (len3 << 8) | len4
	}
}

func (c *CmdParams) writeLength(length int, acc *protocol.PacketPartAccessor) {
	if length>>7 == 0 {
		acc.PutByte(length)

	} else {
		len1 := (length >> 24 & 0xff) | 0x80
		len2 := (length >> 16) & 0xff
		len3 := (length >> 8) & 0xff
		len4 := length & 0xff
		acc.PutBytes([]byte{byte(len1), byte(len2), byte(len3), byte(len4)}, 0, 4)
	}
}

func (c *CmdParams) addParam(name string, value string) {
	c.Params = append(c.Params, []string{name, value})
}

/****************************************/
/* Custome method                       */
/****************************************/
