package base

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	inboundship "bayserver-core/baykit/bayserver/common/inboundship/impl"
	"bayserver-core/baykit/bayserver/common/inboundship/inboundshipstore"
	"bayserver-core/baykit/bayserver/docker"
	common2 "bayserver-core/baykit/bayserver/docker/common"
	"bayserver-core/baykit/bayserver/protocol"
	store2 "bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/ioutil"
	"bayserver-core/baykit/bayserver/util/strutil"
	"net"
	"strconv"
	"strings"
)

type PortSub interface {
	docker.Port
	DockerInitializer
	SupportAnchored() bool
	SupportUnanchored() bool
}

type PortBase struct {
	*DockerBase
	parent            PortSub
	host              string
	port              int
	socketPath        string
	timeoutSec        int
	SecureDocker      docker.Secure
	anchored          bool
	additionalHeaders [][]string
	cities            common2.Cities
	permissionList    []docker.Permission
}

func NewPortBase(parent PortSub) *PortBase {
	p := &PortBase{}
	p.DockerBase = NewDockerBase(parent)
	p.parent = parent
	p.host = ""
	p.port = 0
	p.socketPath = ""
	p.timeoutSec = -1
	p.SecureDocker = nil
	p.anchored = true
	p.additionalHeaders = [][]string{}
	p.cities = common2.NewCities()
	p.permissionList = []docker.Permission{}

	var _ docker.Docker = p // implement check
	return p
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (p *PortBase) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	cerr := p.DockerBase.Init(elm, parent)
	if cerr != nil {
		return cerr
	}

	if elm.Arg == "" {
		return exception2.NewConfigException(
			elm.FileName,
			elm.LineNo,
			baymessage.Get(symbol.CFG_INVALID_PORT_NAME, elm.Name))
	}

	portName := strings.ToLower(elm.Arg)
	if strutil.StartsWith(portName, ":unix:") {
		// unix domain socket

	} else {
		// TCP or UDP port
		var hostPort string
		if strutil.StartsWith(portName, ":tcp:") {
			// TCP server socket
			p.anchored = true
			hostPort = elm.Arg[:5]

		} else if strutil.StartsWith(portName, ":udp") {
			// UDP server socket
			p.anchored = false
			hostPort = elm.Arg[:5]

		} else {
			// default = TCP server socket
			p.anchored = true
			hostPort = elm.Arg
		}

		parts := strings.Split(hostPort, ":")
		var portStr string
		if len(parts) > 1 {
			p.host = parts[0]
			portStr = parts[1]

		} else {
			p.host = ""
			portStr = parts[0]
		}
		var err error
		p.port, err = strconv.Atoi(portStr)
		if err != nil {
			return exception2.NewConfigException(
				elm.FileName,
				elm.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PORT_NAME, elm.Arg))
		}
	}
	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (p *PortBase) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	switch dkr.GetType() {
	case "permission":
		p.permissionList = append(p.permissionList, dkr.(docker.Permission))

	case "city":
		p.cities.Add(dkr.(docker.City))

	case "secure":
		p.SecureDocker = dkr.(docker.Secure)

	default:
		return p.DockerBase.DefaultInitDocker()
	}
	return true, nil
}

func (p *PortBase) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	switch strings.ToLower(kv.Key) {
	default:
		return p.DockerBase.DefaultInitKeyVal(kv)

	case "timeout":
		var err exception.Exception
		p.timeoutSec, err = strutil.ParseInt(kv.Value)
		if err != nil {
			return false, exception2.NewConfigException(
				kv.FileName,
				kv.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE),
				kv.Value)
		}

	case "addheader":
		vals := strings.SplitN(kv.Value, ":", 2)
		if len(vals) < 1 {
			return false,
				exception2.NewConfigException(
					kv.FileName,
					kv.LineNo,
					baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
		}
		p.additionalHeaders = append(p.additionalHeaders, []string{vals[0], vals[1]})
	}
	return true, nil
}

/****************************************/
/* Implements Port                      */
/****************************************/

func (p *PortBase) Host() string {
	return p.host
}

func (p *PortBase) PortNo() int {
	return p.port
}

func (p *PortBase) SocketPath() string {
	return p.socketPath
}

func (p *PortBase) Address() (string, exception.IOException) {
	return "", nil
}

func (p *PortBase) Anchored() bool {
	return p.anchored
}

func (p *PortBase) Secure() bool {
	return p.SecureDocker != nil
}

func (p *PortBase) TimeoutSec() int {
	return p.timeoutSec
}

func (p *PortBase) AdditionalHeaders() [][]string {
	return p.additionalHeaders
}

func (p *PortBase) Cities() []docker.City {
	return p.cities.Cities()
}

func (p *PortBase) FindCity(name string) docker.City {
	return p.cities.FindCity(name)
}

func (p *PortBase) GetSecureConn(conn net.Conn) (net.Conn, exception.IOException) {
	return p.SecureDocker.GetSecureConn(conn)
}

func (p *PortBase) OnConnected(agentId int, rd rudder.Rudder) exception2.HttpException {

	hterr := p.checkAdmitted(rd)
	if hterr != nil {
		return hterr
	}

	sip := getShipStore(agentId).Rent().(*inboundship.InboundShipImpl)
	agt := agent.Get(agentId)

	var tp common.Transporter
	if p.Secure() {
		tp = p.SecureDocker.NewTransporter(agentId, sip)

	} else {
		size, ioerr := ioutil.GetSockRecvBufSize(rd)
		if ioerr != nil {
			size = 8192
		}
		tp = multiplexer.NewPlainTransporter(agt.NetMultiplexer(), sip, true, size, false)
		tp.Init()
	}

	sto := getProtocolHandlerStore(p.parent.Protocol(), agentId)
	protoHnd := sto.Rent().(protocol.ProtocolHandler)
	sip.InitInbound(rd, agentId, tp, p.parent, protoHnd)

	st := common.NewRudderState(rd, tp)
	agt.NetMultiplexer().AddRudderState(rd, st)
	agt.NetMultiplexer().ReqRead(rd)

	return nil
}

func (p *PortBase) ReturnProtocolHandler(agentId int, protoHnd protocol.ProtocolHandler) {
	baylog.Debug("%s Return protocol handler: ", protoHnd)
	getProtocolHandlerStore(protoHnd.Protocol(), agentId).Return(protoHnd, true)
}

func (p *PortBase) ReturnShip(sip ship.Ship) {
	baylog.Debug("%s Return ship: ", sip)
	getShipStore(sip.AgentId()).Return(sip, true)
}

/****************************************/
/* Private methods                      */
/****************************************/

func (p *PortBase) checkAdmitted(rd rudder.Rudder) exception2.HttpException {
	for _, p := range p.permissionList {
		hterr := p.SocketAdmitted(rd)
		if hterr != nil {
			return hterr
		}
	}
	return nil
}

func getShipStore(agentId int) *inboundshipstore.InboundShipStore {
	return inboundshipstore.Get(agentId)
}

func getProtocolHandlerStore(protocol string, agentId int) *store2.ProtocolHandlerStore {
	return store2.GetStore(protocol, true, agentId)
}
