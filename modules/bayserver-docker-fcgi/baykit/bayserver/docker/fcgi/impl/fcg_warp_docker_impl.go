package impl

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-docker-fcgi/baykit/bayserver/docker/fcgi"
	"strings"
)

var warpRegisterd = false

type FcgWarpDockerImpl struct {
	*base.WarpBase
	scriptBase string
	docRoot    string
}

func NewFcgWarpDocker() docker.Docker {
	if !warpRegisterd {
		registerWarpProtocols()
	}
	dkr := &FcgWarpDockerImpl{}
	dkr.WarpBase = base.NewWarpBase(dkr)

	var _ docker.Warp = dkr        // implement check
	var _ docker.Club = dkr        // implement check
	var _ fcgi.FcgWarpDocker = dkr // implement check
	return dkr
}

func (d *FcgWarpDockerImpl) String() string {
	return "FcgWarp"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *FcgWarpDockerImpl) Init(elm *bcf.BcfElement, parent docker.Docker) exception.ConfigException {
	err := d.WarpBase.Init(elm, parent)
	if err != nil {
		return err
	}

	if d.scriptBase == "" {
		baylog.Warn("docRoot is not specified")
	}

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (d *FcgWarpDockerImpl) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception.ConfigException) {
	switch strings.ToLower(kv.Key) {
	case "scriptbase":
		d.scriptBase = kv.Value

	case "docroot":
		d.docRoot = kv.Value

	default:
		return d.WarpBase.InitKeyVal(kv)
	}

	return true, nil
}

/****************************************/
/* Implements WarpSub                  */
/****************************************/

func (d *FcgWarpDockerImpl) Secure() bool {
	return false
}

func (d *FcgWarpDockerImpl) Protocol() string {
	return fcgi.FCG_PROTO_NAME
}

func (d *FcgWarpDockerImpl) NewTransporter(agt agent.GrandAgent, rd rudder.Rudder, sip ship.Ship) (*multiplexer.PlainTransporter, exception2.IOException) {
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
/* Implements FcgWarpDocker             */
/****************************************/

func (d *FcgWarpDockerImpl) ScriptBase() string {
	return d.scriptBase
}

func (d *FcgWarpDockerImpl) DocRoot() string {
	return d.docRoot
}

/****************************************/
/* Private function                      */
/****************************************/

/****************************************/
/* Static function                      */
/****************************************/

func registerWarpProtocols() {
	packetstore.RegisterPacketProtocol(
		fcgi.FCG_PROTO_NAME,
		fcgi.FcgPacketFactory,
	)
	protocolhandlerstore.RegisterProtocol(
		fcgi.FCG_PROTO_NAME,
		false,
		fcgi.FcgWarpProtocolHandlerFactory,
	)
	warpRegisterd = true
}
