package h2

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

const PACKET_STATE_READ_LENGTH = 1
const PACKET_STATE_READ_TYPE = 2
const PACKET_STATE_READ_FLAGS = 3
const PACKET_STATE_READ_STREAM_IDENTIFIER = 4
const PACKET_STATE_READ_FRAME_PAYLOAD = 5
const PACKET_STATE_END = 6

const FRAME_LEN_LENGTH = 3
const FRAME_LEN_TYPE = 1
const FRAME_LEN_FLAGS = 1
const FRAME_LEN_STREAM_IDENTIFIER = 4

var CONNECTION_PREFACE = []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")

type frameHeaderItem struct {
	start  int
	length int
	pos    int // relative reading position
}

func newFrameHeaderItem(start int, length int) *frameHeaderItem {
	return &frameHeaderItem{
		start:  start,
		length: length,
		pos:    0,
	}
}

func (i *frameHeaderItem) Get(buf *bytes.Buffer, index int) int {
	return int(buf.Bytes()[i.start+index]) & 0xFF
}

type H2PacketUnpacker struct {
	impl.PacketUnpackerImpl
	state       int
	tmpBuf      *bytes.Buffer
	item        *frameHeaderItem
	prefaceRead bool
	typ         int
	payloadLen  int
	flags       int
	streamId    int

	cmdUnpacker *H2CommandUnPacker
	pktStore    *packetstore.PacketStore
	serverMode  bool

	contLen   int
	readBytes int
	pos       int
}

func NewH2PacketUnpacker(cmdUnpacker *H2CommandUnPacker, pktStore *packetstore.PacketStore, serverMode bool) *H2PacketUnpacker {
	pu := H2PacketUnpacker{}
	pu.cmdUnpacker = cmdUnpacker
	pu.pktStore = pktStore
	pu.serverMode = serverMode
	pu.tmpBuf = bytes.NewBuffer([]byte{})
	pu.Reset()
	return &pu
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (pu *H2PacketUnpacker) Reset() {
	pu.resetState()
	pu.prefaceRead = false
}

/****************************************/
/* Implements PacketUnpacker            */
/****************************************/

func (pu *H2PacketUnpacker) BytesReceived(buf []byte) (common.NextSocketAction, exception2.IOException) {

	var ioerr exception2.IOException = nil

	for { // try-catch
		suspend := false

		pu.pos = 0
		if pu.serverMode && !pu.prefaceRead {
			length := len(CONNECTION_PREFACE) - pu.tmpBuf.Len()
			if length > len(buf) {
				length = len(buf)
			}
			pu.tmpBuf.Write(buf[pu.pos : pu.pos+length])
			pu.pos += length

			if pu.tmpBuf.Len() == len(CONNECTION_PREFACE) {
				for i := 0; i < pu.tmpBuf.Len(); i++ {
					if CONNECTION_PREFACE[i] != pu.tmpBuf.Bytes()[i] {
						ioerr = exception.NewProtocolException("Invalid connection preface: " + string(pu.tmpBuf.Bytes()[0:pu.tmpBuf.Len()]))
						break
					}
				}
				pkt := pu.pktStore.Rent(H2_TYPE_PREFACE)
				pkt.NewDataAccessor().PutBytes(pu.tmpBuf.Bytes(), 0, pu.tmpBuf.Len())
				var nstat common.NextSocketAction
				nstat, ioerr = pu.cmdUnpacker.PacketReceived(pkt)
				if ioerr != nil {
					break
				}
				pu.pktStore.Return(pkt)
				if nstat != common.NEXT_SOCKET_ACTION_CONTINUE {
					return nstat, nil
				}

				baylog.Debug("Connection preface OK")
				pu.prefaceRead = true
				pu.tmpBuf.Reset()
			}
		}

		for pu.state != PACKET_STATE_END && pu.pos < len(buf) {
			switch pu.state {
			case PACKET_STATE_READ_LENGTH:
				if pu.readHeaderItem(buf) {
					pu.payloadLen = (pu.item.Get(pu.tmpBuf, 0)&0xFF)<<16 |
						(pu.item.Get(pu.tmpBuf, 1)&0xFF)<<8 |
						(pu.item.Get(pu.tmpBuf, 2) & 0xFF)
					pu.item = newFrameHeaderItem(pu.tmpBuf.Len(), FRAME_LEN_TYPE)
					pu.changeState(PACKET_STATE_READ_TYPE)
				}
				break

			case PACKET_STATE_READ_TYPE:
				if pu.readHeaderItem(buf) {
					pu.typ = pu.item.Get(pu.tmpBuf, 0)
					pu.item = newFrameHeaderItem(pu.tmpBuf.Len(), FRAME_LEN_FLAGS)
					pu.changeState(PACKET_STATE_READ_FLAGS)
				}
				break

			case PACKET_STATE_READ_FLAGS:
				if pu.readHeaderItem(buf) {
					pu.flags = pu.item.Get(pu.tmpBuf, 0)
					pu.item = newFrameHeaderItem(pu.tmpBuf.Len(), FRAME_LEN_STREAM_IDENTIFIER)
					pu.changeState(PACKET_STATE_READ_STREAM_IDENTIFIER)
				}
				break

			case PACKET_STATE_READ_STREAM_IDENTIFIER:
				if pu.readHeaderItem(buf) {
					pu.streamId = ((pu.item.Get(pu.tmpBuf, 0) & 0x7F) << 24) |
						(pu.item.Get(pu.tmpBuf, 1) << 16) |
						(pu.item.Get(pu.tmpBuf, 2) << 8) |
						pu.item.Get(pu.tmpBuf, 3)
					pu.item = newFrameHeaderItem(pu.tmpBuf.Len(), pu.payloadLen)
					pu.changeState(PACKET_STATE_READ_FRAME_PAYLOAD)
				}
				break

			case PACKET_STATE_READ_FRAME_PAYLOAD:
				if pu.readHeaderItem(buf) {
					pu.changeState(PACKET_STATE_END)
				}
				break

			default:
				bayserver.FatalError(exception2.NewSink("Illegal state"))
			}

			if pu.state == PACKET_STATE_END {
				pkt := pu.pktStore.Rent(pu.typ).(*H2Packet)
				pkt.StreamId = pu.streamId
				pkt.Flags = NewH2Flags(pu.flags)
				pkt.NewHeaderAccessor().PutBytes(pu.tmpBuf.Bytes(), 0, FRAME_HEADER_LEN)
				pkt.NewDataAccessor().PutBytes(pu.tmpBuf.Bytes(), FRAME_HEADER_LEN, pu.tmpBuf.Len()-FRAME_HEADER_LEN)

				var nxtAct common.NextSocketAction

				nxtAct, ioerr = pu.cmdUnpacker.PacketReceived(pkt)
				/*finally */ {
					pu.pktStore.Return(pkt)
					pu.resetState()
				}

				if ioerr != nil {
					break
				}

				if nxtAct == common.NEXT_SOCKET_ACTION_SUSPEND {
					suspend = true

				} else if nxtAct != common.NEXT_SOCKET_ACTION_CONTINUE {
					return nxtAct, nil
				}
			}
		}

		if suspend {
			return common.NEXT_SOCKET_ACTION_SUSPEND, nil

		} else {
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}
	}

	return -1, ioerr
}

/****************************************/
/* Private functions                    */
/****************************************/

func (pu *H2PacketUnpacker) readHeaderItem(buf []byte) bool {
	length := pu.item.length - pu.item.pos
	if len(buf)-pu.pos < length {
		length = len(buf) - pu.pos
	}
	pu.tmpBuf.Write(buf[pu.pos : pu.pos+length])
	pu.pos += length
	pu.item.pos += length

	return pu.item.pos == pu.item.length
}

func (pu *H2PacketUnpacker) changeState(newState int) {
	pu.state = newState
}

func (pu *H2PacketUnpacker) resetState() {
	pu.changeState(PACKET_STATE_READ_LENGTH)
	pu.item = newFrameHeaderItem(0, FRAME_LEN_LENGTH)
	pu.contLen = 0
	pu.readBytes = 0
	pu.tmpBuf.Reset()
	pu.typ = -1
	pu.flags = 0
	pu.streamId = 0
	pu.payloadLen = 0
}
