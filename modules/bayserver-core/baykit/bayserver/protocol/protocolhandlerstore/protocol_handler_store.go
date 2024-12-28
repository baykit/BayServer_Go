package protocolhandlerstore

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/common/objectstore"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/strutil"
)

var protoMap = map[string]*ProtocolInfo{}

/************************************************/
/* Type ProtocolHandlerStore_LifeCycleListener  */
/************************************************/

type ProtocolHandlerStore_LifeCycleListener struct {
	// implements common.LifecycleListener
}

func (l *ProtocolHandlerStore_LifeCycleListener) Add(agentId int) {
	for _, val := range protoMap {
		val.AddAgent(agentId)
	}
}

func (l *ProtocolHandlerStore_LifeCycleListener) Remove(agentId int) {
	for _, val := range protoMap {
		val.RemoveAgent(agentId)
	}
}

/****************************************/
/* Type ProtocolInfo                     */
/****************************************/

type ProtocolInfo struct {
	protocol               string
	serverMode             bool
	protocolHandlerFactory ProtocolHandlerFactory
	stores                 map[int]*ProtocolHandlerStore
}

func NewProtocolInfo(protocol string, serverMode bool, factory ProtocolHandlerFactory) *ProtocolInfo {
	return &ProtocolInfo{
		protocol:               protocol,
		serverMode:             serverMode,
		protocolHandlerFactory: factory,
		stores:                 map[int]*ProtocolHandlerStore{},
	}
}

func (p *ProtocolInfo) AddAgent(agentId int) {
	store := packetstore.GetStore(p.protocol, agentId)
	p.stores[agentId] = NewProtocolHandlerStore(p.protocol, p.serverMode, p.protocolHandlerFactory, store)
}

func (p *ProtocolInfo) RemoveAgent(agentId int) {
	delete(p.stores, agentId)
}

/****************************************/
/* Type ProtocolHandlerStore            */
/****************************************/

type ProtocolHandlerStore struct {
	objectstore.ObjectStore
	protocol   string
	serverMode bool
}

func NewProtocolHandlerStore(proto string, serverMode bool, fact ProtocolHandlerFactory, pktStore *packetstore.PacketStore) *ProtocolHandlerStore {
	return &ProtocolHandlerStore{
		ObjectStore: *objectstore.NewObjectStore(func() util.Reusable { return fact(pktStore) }),
		protocol:    proto,
		serverMode:  serverMode,
	}
}

func (h *ProtocolHandlerStore) PrintUsage(indent int) {
	var typ string
	if h.serverMode {
		typ = "s"
	} else {
		typ = "c"
	}
	baylog.Info("%sProtocolHandlerStore(%s%s) Usage:", strutil.Indent(indent), h.protocol, typ)
	h.ObjectStore.PrintUsage(indent + 1)

}

/****************************************/
/* Static functions                     */
/****************************************/

func Init() {
	agent.AddLifeCycleListener(&ProtocolHandlerStore_LifeCycleListener{})
}

func GetStore(proto string, serverMode bool, agentId int) *ProtocolHandlerStore {
	return protoMap[constructProtocol(proto, serverMode)].stores[agentId]
}

func GetStores(agentId int) []*ProtocolHandlerStore {
	storeList := make([]*ProtocolHandlerStore, 0)
	for _, ifo := range protoMap {
		storeList = append(storeList, ifo.stores[agentId])
	}
	return storeList
}

func RegisterProtocol(
	protocol string,
	serverMode bool,
	factory ProtocolHandlerFactory) {

	key := constructProtocol(protocol, serverMode)
	if _, contains := protoMap[key]; !contains {
		protoMap[key] = NewProtocolInfo(protocol, serverMode, factory)
	}
}

func constructProtocol(protocol string, serverMode bool) string {
	if serverMode {
		return protocol + "-s"
	} else {
		return protocol + "-c"
	}
}
