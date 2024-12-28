package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bytes"
)

/**
 * AJP Protocol
 * https://tomcat.apache.org/connectors-doc/ajp/ajpv13a.html
 *
 */

const STATE_READ_PREAMBLE = 1 // state for reading first 8 bytes (from version to reserved)
const STATE_READ_BODY = 2     // state for reading content data
const STATE_END = 3

type AjpPacketUnpacker struct {
	impl.PacketUnpackerImpl

	preambleBuf *bytes.Buffer
	bodyBuf     *bytes.Buffer
	state       int
	pktStore    *packetstore.PacketStore
	cmdUnpacker *AjpCommandUnpacker

	bodyLen   int
	readBytes int
	ajpType   int
	toServer  bool
	needData  bool
}

func NewAjpPacketUnpacker(cmdUnpacker *AjpCommandUnpacker, pktStore *packetstore.PacketStore) *AjpPacketUnpacker {
	pu := AjpPacketUnpacker{
		preambleBuf: &bytes.Buffer{},
		bodyBuf:     &bytes.Buffer{},
		state:       STATE_READ_PREAMBLE,
		cmdUnpacker: cmdUnpacker,
		pktStore:    pktStore,
	}
	return &pu
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (pu *AjpPacketUnpacker) Reset() {
	pu.state = STATE_READ_PREAMBLE
	pu.ajpType = -1
	pu.bodyLen = 0
	pu.needData = false
	pu.preambleBuf.Reset()
	pu.bodyBuf.Reset()
}

/****************************************/
/* Implements PacketUnpacker            */
/****************************************/

func (pu *AjpPacketUnpacker) BytesReceived(buf []byte) (common.NextSocketAction, exception2.IOException) {
	suspend := false
	pos := 0

	//baylog.Debug("BytesReceived: len=%d", len(buf))

	for pos < len(buf) {
		if pu.state == STATE_READ_PREAMBLE {
			length := AJP_PREAMBLE_SIZE - pu.preambleBuf.Len()
			if len(buf)-pos < length {
				length = len(buf) - pos
			}
			pu.preambleBuf.Write(buf[pos : pos+length])
			pos += length

			if pu.preambleBuf.Len() == AJP_PREAMBLE_SIZE {
				ioerr := pu.preambleRead()
				if ioerr != nil {
					return -1, ioerr
				}
				pu.changeState(STATE_READ_BODY)
			}
		}

		if pu.state == STATE_READ_BODY {
			length := pu.bodyLen - pu.bodyBuf.Len()
			if length > len(buf)-pos {
				length = len(buf) - pos
			}

			pu.bodyBuf.Write(buf[pos : pos+length])
			pos += length

			if pu.bodyBuf.Len() == pu.bodyLen {
				pu.bodyRead()
				pu.changeState(STATE_END)
			}
		}

		if pu.state == STATE_END {
			baylog.Info("buflen=%d pos=%d", len(buf), pos)
			pkt := pu.pktStore.Rent(pu.ajpType)
			pkt.(*AjpPacket).toServer = pu.toServer
			pkt.(*AjpPacket).NewAjpHeaderAccessor().PutBytes(pu.preambleBuf.Bytes(), 0, pu.preambleBuf.Len())
			pkt.(*AjpPacket).NewDataAccessor().PutBytes(pu.bodyBuf.Bytes(), 0, pu.bodyBuf.Len())

			nextAct, ioerr := pu.cmdUnpacker.PacketReceived(pkt)
			pu.pktStore.Return(pkt)
			if ioerr != nil {
				return -1, ioerr
			}

			pu.Reset()
			pu.needData = pu.cmdUnpacker.NeedData()

			if nextAct == common.NEXT_SOCKET_ACTION_SUSPEND {
				suspend = true

			} else if nextAct != common.NEXT_SOCKET_ACTION_CONTINUE {
				return nextAct, nil
			}
		}
	}

	if suspend {
		return common.NEXT_SOCKET_ACTION_SUSPEND, nil

	} else {
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}
}

/****************************************/
/* Private functions                    */
/****************************************/

func (pu *AjpPacketUnpacker) changeState(newState int) {
	pu.state = newState
}

func (pu *AjpPacketUnpacker) preambleRead() exception2.IOException {
	data := pu.preambleBuf.Bytes()

	if data[0] == 0x12 && data[1] == 0x34 {
		pu.toServer = true

	} else if data[0] == 'A' && data[1] == 'B' {
		pu.toServer = false

	} else {
		return exception.NewProtocolException("Must be start with 0x1234 or 'AB'")
	}

	pu.bodyLen = (int(data[2]) << 8) | (int(data[3])&0xff)&0xffff
	return nil
}

func (pu *AjpPacketUnpacker) bodyRead() {
	if pu.needData {
		pu.ajpType = AJP_TYPE_DATA

	} else {
		pu.ajpType = int(pu.bodyBuf.Bytes()[0] & 0xff)
	}
}
