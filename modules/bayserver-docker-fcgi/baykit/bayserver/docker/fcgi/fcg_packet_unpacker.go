package fcgi

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bytes"
)

/**
 * Packet unmarshall logic for FCGI
 */

const PACKET_STATE_READ_PREAMBLE = 1 // state for reading first 8 bytes (from version to reserved)
const PACKET_STATE_READ_CONTENT = 2  // state for reading content data
const PACKET_STATE_READ_PADDING = 3  // state for reading padding data
const PACKET_STATE_END = 4

type FcgPacketUnpacker struct {
	impl.PacketUnpackerImpl

	headerBuf        *bytes.Buffer
	dataBuf          *bytes.Buffer
	version          int
	fcgType          int
	reqId            int
	length           int
	padding          int
	paddingReadBytes int
	state            int

	cmdUnpacker *FcgCommandUnpacker
	pktStore    *packetstore.PacketStore
	contLen     int
	readBytes   int
}

func NewFcgPacketUnpacker(cmdUnpacker *FcgCommandUnpacker, pktStore *packetstore.PacketStore) *FcgPacketUnpacker {
	pu := FcgPacketUnpacker{
		headerBuf:   &bytes.Buffer{},
		dataBuf:     &bytes.Buffer{},
		cmdUnpacker: cmdUnpacker,
		pktStore:    pktStore,
	}
	pu.Reset()
	return &pu
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (pu *FcgPacketUnpacker) Reset() {
	pu.state = PACKET_STATE_READ_PREAMBLE
	pu.version = 0
	pu.fcgType = -1
	pu.reqId = 0
	pu.length = 0
	pu.padding = 0
	pu.paddingReadBytes = 0
	pu.contLen = 0
	pu.readBytes = 0
	pu.headerBuf.Reset()
	pu.dataBuf.Reset()
}

/****************************************/
/* Implements PacketUnpacker            */
/****************************************/

func (pu *FcgPacketUnpacker) BytesReceived(buf []byte) (common.NextSocketAction, exception2.IOException) {
	nextSuspend := false
	nextWrite := false
	pos := 0

	//baylog.Debug("BytesReceived: len=%d", len(buf))

	for pos < len(buf) {
		for pu.state != PACKET_STATE_END && pos < len(buf) {

			switch pu.state {

			case PACKET_STATE_READ_PREAMBLE:
				// preamble read mode
				length := FCG_PREAMBLE_SIZE - pu.headerBuf.Len()
				if len(buf)-pos < length {
					length = len(buf) - pos
				}

				pu.headerBuf.Write(buf[pos : pos+length])
				pos += length
				if pu.headerBuf.Len() == FCG_PREAMBLE_SIZE {
					pu.headerReadDone()
					if pu.length == 0 {
						if pu.padding == 0 {
							pu.changeState(PACKET_STATE_END)
						} else {
							pu.changeState(PACKET_STATE_READ_PADDING)
						}

					} else {
						pu.changeState(PACKET_STATE_READ_CONTENT)
					}
				}

			case PACKET_STATE_READ_CONTENT:
				// content read mode
				length := pu.length - pu.dataBuf.Len()
				if length > len(buf)-pos {
					length = len(buf) - pos
				}

				if length > 0 {
					pu.dataBuf.Write(buf[pos : pos+length])
					pos += length

					if pu.dataBuf.Len() == pu.length {
						if pu.padding == 0 {
							pu.changeState(PACKET_STATE_END)
						} else {
							pu.changeState(PACKET_STATE_READ_PADDING)
						}
					}
				}

			case PACKET_STATE_READ_PADDING:
				// padding read mode
				length := pu.padding - pu.paddingReadBytes

				if length > len(buf)-pos {
					length = len(buf) - pos
				}

				pos += length

				if length > 0 {
					pu.paddingReadBytes += length

					if pu.paddingReadBytes == pu.padding {
						pu.changeState(PACKET_STATE_END)
					}
				}

			default:
				bayserver.FatalError(exception2.NewSink("Illegal State"))
			}
		}

		if pu.state == PACKET_STATE_END {
			pkt := pu.pktStore.Rent(pu.fcgType)
			pkt.(*FcgPacket).ReqId = pu.reqId
			pkt.NewHeaderAccessor().PutBytes(pu.headerBuf.Bytes(), 0, pu.headerBuf.Len())
			pkt.NewDataAccessor().PutBytes(pu.dataBuf.Bytes(), 0, pu.dataBuf.Len())

			nextAct, ioerr := pu.cmdUnpacker.PacketReceived(pkt)
			pu.pktStore.Return(pkt)
			if ioerr != nil {
				return -1, ioerr
			}

			pu.Reset()

			switch nextAct {
			case common.NEXT_SOCKET_ACTION_SUSPEND:
				nextSuspend = true
			case common.NEXT_SOCKET_ACTION_CONTINUE:
				break
			case common.NEXT_SOCKET_ACTION_WRITE:
				nextWrite = true
			case common.NEXT_SOCKET_ACTION_CLOSE:
				return nextAct, nil
			}
		}
	}

	if nextWrite {
		return common.NEXT_SOCKET_ACTION_WRITE, nil

	} else if nextSuspend {
		return common.NEXT_SOCKET_ACTION_SUSPEND, nil

	} else {
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}
}

/****************************************/
/* Private functions                    */
/****************************************/

func (pu *FcgPacketUnpacker) changeState(newState int) {
	pu.state = newState
}

func (pu *FcgPacketUnpacker) headerReadDone() {
	pre := pu.headerBuf.Bytes()
	pu.version = byteToInt(pre[0])
	pu.fcgType = byteToInt(pre[1])
	pu.reqId = bytesToInt(pre[2], pre[3])
	pu.length = bytesToInt(pre[4], pre[5])
	pu.padding = byteToInt(pre[6])
	_ = byteToInt(pre[7])

	baylog.Debug("fcg: read packet header: version=%s type=%d reqId=%d length=%d padding=%d",
		pu.version, pu.fcgType, pu.reqId, pu.length, pu.padding)
}

func byteToInt(b byte) int {
	return int(b) & 0xff
}

func bytesToInt(b1 byte, b2 byte) int {
	return byteToInt(b1)<<8 | byteToInt(b2)
}
