package h1

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bytes"
)

/**
 * Read HTTP header
 *
 *   HTTP/1.x has no packet format. So we make HTTP header and content pretend to be packet
 *
 *   From RFC2616
 *   generic-message : start-line
 *                     (message-header CRLF)*
 *                     CRLF
 *                     [message-body]
 *
 *
 */

const PACKET_STATE_READ_HEADERS = 1
const PACKET_STATE_READ_CONTENT = 2
const PACKET_STATE_END = 3

const MAX_LINE_LEN = 8193

type H1PacketUnpacker struct {
	impl.PacketUnpackerImpl
	state       int
	cmdUnpacker *H1CommandUnpacker
	pktStore    *packetstore.PacketStore
	tmpBuf      *bytes.Buffer
}

func NewH1PacketUnpacker(cmdUnpacker *H1CommandUnpacker, pktStore *packetstore.PacketStore) *H1PacketUnpacker {
	pu := H1PacketUnpacker{}
	pu.cmdUnpacker = cmdUnpacker
	pu.pktStore = pktStore
	pu.tmpBuf = bytes.NewBuffer([]byte{})
	pu.resetState()
	return &pu
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (pu *H1PacketUnpacker) Reset() {
	pu.resetState()
}

/****************************************/
/* Implements PacketUnpacker            */
/****************************************/

func (pu *H1PacketUnpacker) BytesReceived(buf []byte) (common.NextSocketAction, exception2.IOException) {
	if pu.state == PACKET_STATE_END {
		pu.Reset()
		bayserver.FatalError(exception2.NewSink("Illegal State"))
	}

	pos := 0
	lineLen := 0
	suspend := false

	if pu.state == PACKET_STATE_READ_HEADERS {
		for pos < len(buf) {
			b := buf[pos]
			pu.tmpBuf.WriteByte(b)
			pos++
			if b == '\r' {
				continue

			} else if b == '\n' {
				if lineLen == 0 {
					pkt := pu.pktStore.Rent(H1_TYPE_HEADER)
					pkt.NewDataAccessor().PutBytes(pu.tmpBuf.Bytes(), 0, pu.tmpBuf.Len())
					nextAct, ioerr := pu.cmdUnpacker.PacketReceived(pkt)
					pu.pktStore.Return(pkt)
					if ioerr != nil {
						return -1, ioerr
					}

					switch nextAct {
					case common.NEXT_SOCKET_ACTION_CONTINUE, common.NEXT_SOCKET_ACTION_SUSPEND:
						if pu.cmdUnpacker.ReqFinished() {
							pu.changeState(PACKET_STATE_END)
						} else {
							pu.changeState(PACKET_STATE_READ_CONTENT)
						}

					case common.NEXT_SOCKET_ACTION_CLOSE:
						// Maybe Error
						pu.resetState()
						return nextAct, nil
					}

					suspend = nextAct == common.NEXT_SOCKET_ACTION_SUSPEND
					break
				}
				lineLen = 0

			} else {
				lineLen++
			}

			if lineLen >= MAX_LINE_LEN {
				return -1, exception.NewProtocolException("HTTP/1 Line is too long")
			}
		}
	}

	if pu.state == PACKET_STATE_READ_CONTENT {
		for pos < len(buf) {
			pkt := pu.pktStore.Rent(H1_TYPE_CONTENT)
			length := len(buf) - pos
			if length > MAX_DATA_LEN {
				length = MAX_DATA_LEN
			}

			pkt.NewDataAccessor().PutBytes(buf, pos, length)
			pos += length

			nextAct, ioerr := pu.cmdUnpacker.PacketReceived(pkt)
			pu.pktStore.Return(pkt)

			if ioerr != nil {
				return -1, ioerr
			}

			switch nextAct {
			case common.NEXT_SOCKET_ACTION_CONTINUE, common.NEXT_SOCKET_ACTION_WRITE:
				if pu.cmdUnpacker.ReqFinished() {
					pu.changeState(PACKET_STATE_END)
				}

			case common.NEXT_SOCKET_ACTION_SUSPEND:
				suspend = true

			case common.NEXT_SOCKET_ACTION_CLOSE:
				pu.resetState()
				return nextAct, nil
			}
		}
	}

	if pu.state == PACKET_STATE_END {
		pu.resetState()
	}

	if suspend {
		baylog.Debug("H1 read suspend")
		return common.NEXT_SOCKET_ACTION_SUSPEND, nil

	} else {
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}
}

/****************************************/
/* Private functions                    */
/****************************************/

func (pu *H1PacketUnpacker) changeState(newState int) {
	pu.state = newState
}

func (pu *H1PacketUnpacker) resetState() {
	pu.changeState(PACKET_STATE_READ_HEADERS)
	pu.tmpBuf.Reset()
}
