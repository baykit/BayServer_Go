package base

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/warpship"
	"bayserver-core/baykit/bayserver/common/warpship/warpshipstore"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"net"
	"strconv"
	"strings"
	"sync"
)

/****************************************/
/*  Type MemUsage_LifeCycleListener     */
/****************************************/

type WarpBase_LifeCycleListener struct {
	// implements common.LifecycleListener

	base *WarpBase
}

func NewWarpBase_LifeCycleListener(base *WarpBase) *WarpBase_LifeCycleListener {
	lis := &WarpBase_LifeCycleListener{
		base: base,
	}

	// interface check
	var _ common.LifecycleListener = lis
	return lis
}

func (l *WarpBase_LifeCycleListener) Add(agentId int) {
	l.base.stores[agentId] = warpshipstore.NewWarpShipStore(l.base.maxShips)
}

func (l *WarpBase_LifeCycleListener) Remove(agentId int) {
	delete(l.base.stores, agentId)
}

/****************************************/
/*  Type MemUsage_LifeCycleListener     */
/****************************************/

type WarpSub interface {
	Secure() bool
	Protocol() string
	NewTransporter(agt agent.GrandAgent, rd rudder.Rudder, sip ship.Ship) (*multiplexer.PlainTransporter, exception.IOException)
}

type WarpBase struct {
	*ClubBase

	sub        WarpSub
	scheme     string
	host       string
	port       int
	destTown   string
	maxShips   int
	hostAddr   net.Addr
	timeoutSec int

	tourList     []tour.Tour
	tourListLock sync.Mutex

	/** Agent ID => WarpShipStore */
	stores map[int]*warpshipstore.WarpShipStore
}

func NewWarpBase(sub WarpSub) *WarpBase {
	h := &WarpBase{
		sub:      sub,
		tourList: make([]tour.Tour, 0),
		stores:   make(map[int]*warpshipstore.WarpShipStore),
	}
	h.ClubBase = NewClubBase(sub.(DockerInitializer))

	var _ docker.Warp = h // implement check
	var _ docker.Club = h // implement check
	return h
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (h *WarpBase) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	cerr := h.ClubBase.Init(elm, parent)
	if cerr != nil {
		return cerr
	}

	if h.destTown == "" {
		h.destTown = "/"
	}

	var ioerr exception.IOException = nil
	for { // try catch
		if h.host != "" && strings.HasPrefix(h.host, ":unix:") {
			//sktPath := h.host[6:]
			//h.hostAddr = sysutil.GetUnixDomainSocketAddress(sktPath)
			h.port = -1

		} else {
			if h.port < 0 {
				h.port = 80
			}

			var err error
			h.hostAddr, err = net.ResolveTCPAddr("tcp", h.host+":"+strconv.Itoa(h.port))
			if err != nil {
				ioerr = exception.NewIOExceptionFromError(err)
				break
			}
		}
		break
	}

	if ioerr != nil {
		return exception2.NewConfigException(elm.FileName, elm.LineNo, baymessage.Get(symbol.CFG_INVALID_WARP_DESTINATION, h.host), ioerr)
	}

	agent.AddLifeCycleListener(NewWarpBase_LifeCycleListener(h))
	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (h *WarpBase) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	return false, nil
}

func (h *WarpBase) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	var err error = nil
	switch strings.ToLower(kv.Key) {
	default:
		var cerr exception2.ConfigException
		_, cerr = h.ClubBase.InitKeyVal(kv)
		if cerr != nil {
			return false, cerr
		}

	case "destcity":
		h.host = kv.Value

	case "destport":
		h.port, err = strconv.Atoi(kv.Value)

	case "desttown":
		h.destTown = kv.Value
		if !strings.HasSuffix(h.destTown, "/") {
			h.destTown += "/"
		}

	case "maxships":
		h.maxShips, err = strconv.Atoi(kv.Value)

	case "timeout":
		h.timeoutSec, err = strconv.Atoi(kv.Value)
	}

	if err != nil {
		return false, exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE))
	}
	return true, nil
}

/****************************************/
/* Implements Club                      */
/****************************************/

func (h *WarpBase) Arrive(tur tour.Tour) exception2.HttpException {
	agt := agent.Get(tur.Ship().(ship.Ship).AgentId())
	sto := h.GetShipStore(agt.AgentId())

	wsip := sto.Rent()
	if wsip == nil {
		return exception2.NewHttpException(httpstatus.SERVICE_UNAVAILABLE, "WarpDocker busy")
	}

	var ioerr exception.IOException = nil
	for { // try catch
		needConnect := false
		var tp common.Transporter = nil

		if !wsip.Initialized() {

			rd := impl.NewTcpConnRudderUnconnected()

			tp, ioerr = h.sub.NewTransporter(agt, rd, wsip)
			if ioerr != nil {
				break
			}

			protoHnd := protocolhandlerstore.GetStore(h.sub.Protocol(), false, agt.AgentId()).Rent().(protocol.ProtocolHandler)
			wsip.InitWarp(rd, agt.AgentId(), tp, h.sub.(docker.Warp), protoHnd)

			baylog.Debug("%s init warp ship rd=%s", wsip, rd)
			needConnect = true
		}

		h.tourListLock.Lock()
		h.tourList = append(h.tourList, tur)
		h.tourListLock.Unlock()

		ioerr = wsip.StartWarpTour(tur)
		if ioerr != nil {
			break
		}

		if needConnect {
			agt.NetMultiplexer().AddRudderState(wsip.Rudder(), common.NewRudderState(wsip.Rudder(), tp))
			ioerr = agt.NetMultiplexer().GetTransporter(wsip.Rudder()).ReqConnect(wsip.Rudder(), h.hostAddr)
			if ioerr != nil {
				break
			}
		}

		return nil
	}

	// IOError
	baylog.ErrorE(ioerr, "")
	return exception2.NewHttpException(httpstatus.INTERNAL_SERVER_ERROR, ioerr.Error())
}

/****************************************/
/* Implements Warp                      */
/****************************************/

func (h *WarpBase) Host() string {
	return h.host
}

func (h *WarpBase) Port() int {
	return h.port
}

func (h *WarpBase) DestTown() string {
	return h.destTown
}

func (h *WarpBase) TimeoutSec() int {
	return h.timeoutSec
}

func (h *WarpBase) Keep(wsip ship.Ship) {
	baylog.Debug("%s keep warp ship: %s", h, wsip)
	h.GetShipStore(wsip.AgentId()).Keep(wsip.(warpship.WarpShip))
}

func (h *WarpBase) OnEndShip(wsip ship.Ship) {
	baylog.Debug("%s Return protocol handler: ", wsip)
	h.GetProtocolHandlerStore(wsip.AgentId()).Return(wsip.(warpship.WarpShip).ProtocolHandler(), true)
	baylog.Debug("%s return warp ship", wsip)
	h.GetShipStore(wsip.AgentId()).Return(wsip.(warpship.WarpShip))
}

/****************************************/
/* Custom Functions                     */
/****************************************/

func (h *WarpBase) GetShipStore(agtId int) *warpshipstore.WarpShipStore {
	return h.stores[agtId]
}

func (h *WarpBase) GetProtocolHandlerStore(agtId int) *protocolhandlerstore.ProtocolHandlerStore {
	return protocolhandlerstore.GetStore(h.sub.Protocol(), false, agtId)
}
