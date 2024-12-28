package impl

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/strutil"
	"bayserver-docker-http/baykit/bayserver/docker/http"
	"bayserver-docker-http/baykit/bayserver/docker/http/h1"
	"bayserver-docker-http/baykit/bayserver/docker/http/h2"
	"strings"
)

var warpRegisterd = false

type HtpWarpDockerImpl struct {
	*base.WarpBase

	secure    bool
	supportH2 bool
	traceSsl  bool
}

func NewHtpWarpDocker() docker.Docker {
	if !warpRegisterd {
		registerWarpProtocols()
	}
	dkr := &HtpWarpDockerImpl{}
	dkr.WarpBase = base.NewWarpBase(dkr)

	var _ docker.Docker = dkr // implement check
	var _ docker.Warp = dkr   // implement check
	var _ docker.Club = dkr   // implement check
	var _ base.WarpSub = dkr  // implement check
	var _ http.HtpWarpDocker = &dkr
	return dkr
}

func (d *HtpWarpDockerImpl) String() string {
	return "HtpWarp"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *HtpWarpDockerImpl) Init(elm *bcf.BcfElement, parent docker.Docker) exception.ConfigException {

	err := d.WarpBase.Init(elm, parent)
	if err != nil {
		return err
	}

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (d *HtpWarpDockerImpl) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception.ConfigException) {

	var err exception2.Exception = nil
	switch strings.ToLower(kv.Key) {
	case "supportH2":
		d.supportH2, err = strutil.ParseBool(kv.Value)

	case "tracessl":
		d.traceSsl, err = strutil.ParseBool(kv.Value)

	case "secure":
		d.secure, err = strutil.ParseBool(kv.Value)

	default:
		return d.WarpBase.InitKeyVal(kv)
	}

	if err != nil {
		baylog.ErrorE(err, "")
		return false, exception.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
	}

	return true, nil
}

/****************************************/
/* Implements WarpSub                  */
/****************************************/

func (d *HtpWarpDockerImpl) Secure() bool {
	return false
}

func (d *HtpWarpDockerImpl) Protocol() string {
	return http.H1_PROTO_NAME
}

func (d *HtpWarpDockerImpl) NewTransporter(agt agent.GrandAgent, rd rudder.Rudder, sip ship.Ship) (*multiplexer.PlainTransporter, exception2.IOException) {
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
		http.H1_PROTO_NAME,
		h1.H1PacketFactory,
	)
	packetstore.RegisterPacketProtocol(
		http.H2_PROTO_NAME,
		h2.H2PacketFactory,
	)
	protocolhandlerstore.RegisterProtocol(
		http.H1_PROTO_NAME,
		false,
		h1.H1WarpProtocolHandlerFactory,
	)
	warpRegisterd = true
}
