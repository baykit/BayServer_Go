package impl

import (
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-docker-ajp/baykit/bayserver/docker/ajp"
)

var portRegistered = false

type AjpPortDockerImpl struct {
	*base.PortBase
}

func NewAjpPort() docker.Port {
	if !portRegistered {
		registerProtocols()
	}
	h := &AjpPortDockerImpl{}
	h.PortBase = base.NewPortBase(h)

	// interface check
	var _ docker.Docker = h
	var _ docker.Port = h
	var _ base.PortSub = h
	var _ ajp.AjpPortDocker = h
	return h
}

func (d *AjpPortDockerImpl) String() string {
	return "AjpPortDocker"
}

/****************************************/
/* Implements Port                      */
/****************************************/

func (d *AjpPortDockerImpl) Protocol() string {
	return ajp.AJP_PROTO_NAME
}

/****************************************/
/* Implements PortSub                  */
/****************************************/

func (d *AjpPortDockerImpl) SupportAnchored() bool {
	return true
}

func (d *AjpPortDockerImpl) SupportUnanchored() bool {
	return false
}

/****************************************/
/* Static function                      */
/****************************************/

func registerProtocols() {
	packetstore.RegisterPacketProtocol(
		ajp.AJP_PROTO_NAME,
		ajp.AjpPacketFactory,
	)
	protocolhandlerstore.RegisterProtocol(
		ajp.AJP_PROTO_NAME,
		true,
		ajp.AjpInboundProtocolHandlerFactory,
	)
	portRegistered = true
}
