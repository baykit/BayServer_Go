package impl

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util/exception"
)

type CommandPackerImpl struct {
	pktPacker protocol.PacketPacker
	pktStore  *packetstore.PacketStore
}

func NewCommandPacker(pktPacker protocol.PacketPacker, pktStore *packetstore.PacketStore) *CommandPackerImpl {
	return &CommandPackerImpl{
		pktPacker: pktPacker,
		pktStore:  pktStore,
	}
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (cp *CommandPackerImpl) Reset() {

}

/****************************************/
/* Implements CommandPacker             */
/****************************************/

func (cp *CommandPackerImpl) Post(sip ship.Ship, c protocol.Command, listener common.DataConsumeListener) exception.IOException {
	pkt := cp.pktStore.Rent(c.Type())

	var ioerr exception.IOException = nil
	for { // Try catch
		ioerr = c.Pack(pkt)
		if ioerr != nil {
			break
		}

		ioerr = cp.pktPacker.Post(sip, pkt, func() {
			cp.pktStore.Return(pkt)
			if listener != nil {
				listener()
			}
		})
		if ioerr != nil {
			break
		}

		break
	}

	if ioerr != nil {
		cp.pktStore.Return(pkt)
		return ioerr
	}

	return nil
}
