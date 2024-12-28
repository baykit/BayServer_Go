package impl

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util/exception"
)

type ProtocolHandlerImpl struct {
	protocol.ProtocolHandler
	packetUnpacker  protocol.PacketUnpacker
	packetPacker    protocol.PacketPacker
	commandUnpacker protocol.CommandUnpacker
	commandPacker   protocol.CommandPacker
	commandHandler  protocol.CommandHandler
	ship            ship.Ship
	serverMode      bool
}

func (ph *ProtocolHandlerImpl) ConstructProtocolHandler(
	packetUnpacker protocol.PacketUnpacker,
	packetPacker protocol.PacketPacker,
	commandUnpacker protocol.CommandUnpacker,
	commandPacker protocol.CommandPacker,
	commandHandler protocol.CommandHandler,
	serverMode bool) {
	ph.packetUnpacker = packetUnpacker
	ph.packetPacker = packetPacker
	ph.commandUnpacker = commandUnpacker
	ph.commandPacker = commandPacker
	ph.commandHandler = commandHandler
	ph.serverMode = serverMode
}

func (ph *ProtocolHandlerImpl) String() string {
	return "PH"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (ph *ProtocolHandlerImpl) Reset() {
	ph.commandUnpacker.Reset()
	ph.commandPacker.Reset()
	ph.packetUnpacker.Reset()
	ph.packetUnpacker.Reset()
	ph.commandHandler.Reset()
	ph.ship = nil
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (ph *ProtocolHandlerImpl) PacketUnpacker() protocol.PacketUnpacker {
	return ph.packetUnpacker
}

func (ph *ProtocolHandlerImpl) PacketPacker() protocol.PacketPacker {
	return ph.packetPacker
}

func (ph *ProtocolHandlerImpl) CommandUnpacker() protocol.CommandUnpacker {
	return ph.commandUnpacker
}

func (ph *ProtocolHandlerImpl) CommandPacker() protocol.CommandPacker {
	return ph.commandPacker
}

func (ph *ProtocolHandlerImpl) CommandHandler() protocol.CommandHandler {
	return ph.commandHandler
}

func (ph *ProtocolHandlerImpl) Ship() ship.Ship {
	return ph.ship
}

func (ph *ProtocolHandlerImpl) Init(ship ship.Ship) {
	ph.ship = ship
}

func (ph *ProtocolHandlerImpl) BytesReceived(buf []byte) (common.NextSocketAction, exception.IOException) {
	return ph.packetUnpacker.BytesReceived(buf)
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (ph *ProtocolHandlerImpl) Post(cmd protocol.Command, listener common.DataConsumeListener) exception.IOException {
	return ph.CommandPacker().Post(ph.ship, cmd, listener)
}
