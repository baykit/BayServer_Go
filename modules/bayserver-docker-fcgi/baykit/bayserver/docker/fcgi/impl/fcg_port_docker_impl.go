package impl

import (
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-docker-fcgi/baykit/bayserver/docker/fcgi"
)

var registerd = false

type FcgPortDockerImpl struct {
	*base.PortBase
}

func NewFcgPort() docker.Port {
	if !registerd {
		registerProtocols()
	}
	h := &FcgPortDockerImpl{}
	h.PortBase = base.NewPortBase(h)

	// interface check
	var _ docker.Docker = h
	var _ docker.Port = h
	var _ base.PortSub = h
	var _ fcgi.FcgPortDocker = h
	return h
}

func (d *FcgPortDockerImpl) String() string {
	return "FcgPortDocker"
}

/****************************************/
/* Implements Port                      */
/****************************************/

func (d *FcgPortDockerImpl) Protocol() string {
	return fcgi.FCG_PROTO_NAME
}

/****************************************/
/* Implements PortSub                  */
/****************************************/

func (d *FcgPortDockerImpl) SupportAnchored() bool {
	return true
}

func (d *FcgPortDockerImpl) SupportUnanchored() bool {
	return false
}

/****************************************/
/* Static function                      */
/****************************************/

func registerProtocols() {
	packetstore.RegisterPacketProtocol(
		fcgi.FCG_PROTO_NAME,
		fcgi.FcgPacketFactory,
	)
	protocolhandlerstore.RegisterProtocol(
		fcgi.FCG_PROTO_NAME,
		true,
		fcgi.FcgInboundProtocolHandlerFactory,
	)
	registerd = true
}
