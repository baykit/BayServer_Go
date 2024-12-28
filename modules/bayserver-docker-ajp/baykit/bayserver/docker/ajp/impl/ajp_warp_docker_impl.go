package impl

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/ship"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-docker-ajp/baykit/bayserver/docker/ajp"
)

var warpRegisterd = false

type AjpWarpDockerImpl struct {
	*base.WarpBase
}

func NewAjpWarpDocker() docker.Docker {
	if !warpRegisterd {
		registerWarpProtocols()
	}
	dkr := &AjpWarpDockerImpl{}
	dkr.WarpBase = base.NewWarpBase(dkr)

	var _ docker.Docker = dkr // implement check
	var _ docker.Warp = dkr   // implement check
	var _ docker.Club = dkr   // implement check
	return dkr
}

func (d *AjpWarpDockerImpl) String() string {
	return "AjpWarp"
}

/****************************************/
/* Implements WarpSub                  */
/****************************************/

func (d *AjpWarpDockerImpl) Secure() bool {
	return false
}

func (d *AjpWarpDockerImpl) Protocol() string {
	return ajp.AJP_PROTO_NAME
}

func (d *AjpWarpDockerImpl) NewTransporter(agt agent.GrandAgent, rd rudder.Rudder, sip ship.Ship) (*multiplexer.PlainTransporter, exception2.IOException) {
	bufSize, ioerr := rd.(*impl.TcpConnRudder).GetSocketReceiveBufferSize()
	if ioerr != nil {
		return nil, ioerr
	}
	tp := multiplexer.NewPlainTransporter(
		agt.NetMultiplexer(),
		sip,
		false,
		bufSize,
		false)
	tp.Init()
	return tp, nil
}

/****************************************/
/* Private function                      */
/****************************************/

/****************************************/
/* Static function                      */
/****************************************/

func registerWarpProtocols() {
	packetstore.RegisterPacketProtocol(
		ajp.AJP_PROTO_NAME,
		ajp.AjpPacketFactory,
	)
	protocolhandlerstore.RegisterProtocol(
		ajp.AJP_PROTO_NAME,
		false,
		ajp.AjpWarpProtocolHandlerFactory,
	)
	warpRegisterd = true
}
