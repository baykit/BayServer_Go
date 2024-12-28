package packetstore

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/objectstore"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/strutil"
)

var packetProtocolMap = map[string]*PacketProtocolInfo{}

/****************************************/
/*  Type PacketStore_LifeCycleListener  */
/****************************************/

type PacketStore_LifeCycleListener struct {
	common.LifecycleListener
}

func (l *PacketStore_LifeCycleListener) Add(agentId int) {
	for _, info := range packetProtocolMap {
		info.AddAgent(agentId)
	}
}

func (l *PacketStore_LifeCycleListener) Remove(agentId int) {
	for _, info := range packetProtocolMap {
		info.RemoveAgent(agentId)
	}
}

/****************************************/
/* Type PacketProtocolInfo              */
/****************************************/

type PacketProtocolInfo struct {
	protocol string
	factory  protocol.PacketFactory

	/** Agent ID => PacketStore */
	stores map[int]*PacketStore
}

func NewPacketProtocolInfo(protocol string, factory protocol.PacketFactory) *PacketProtocolInfo {
	return &PacketProtocolInfo{
		protocol: protocol,
		factory:  factory,
		stores:   map[int]*PacketStore{},
	}
}

func (p *PacketProtocolInfo) AddAgent(agentId int) {
	store := NewPacketStore(p.protocol, p.factory)
	p.stores[agentId] = store
}

func (p *PacketProtocolInfo) RemoveAgent(agentId int) {
	delete(p.stores, agentId)
}

/****************************************/
/* Type PacketStore                     */
/****************************************/

type PacketStore struct {
	protocol string
	storeMap map[int]*objectstore.ObjectStore
	factory  protocol.PacketFactory
}

func NewPacketStore(protocol string, factory protocol.PacketFactory) *PacketStore {
	return &PacketStore{
		protocol: protocol,
		storeMap: map[int]*objectstore.ObjectStore{},
		factory:  factory,
	}
}

func (st *PacketStore) Rent(typ int) protocol.Packet {
	store := st.storeMap[typ]
	if store == nil {
		store = objectstore.NewObjectStore(func() util.Reusable { return st.factory(typ) })
		st.storeMap[typ] = store
	}
	// 他の処理を追加
	return store.Rent().(protocol.Packet) // 例: オブジェクトの取得
}

func (st *PacketStore) Return(pkt protocol.Packet) {
	store := st.storeMap[pkt.Type()]
	store.Return(pkt, true)
}

func (st *PacketStore) PrintUsage(indent int) {
	baylog.Info("%sPacketStore(%s) usage nTypes=%d", strutil.Indent(indent), st.protocol, len(st.storeMap))
	for typ := range st.storeMap {
		baylog.Info("%sType: %d", strutil.Indent(indent+1), typ)
		st.storeMap[typ].PrintUsage(indent + 2)
	}
}

/****************************************/
/* Static functions                     */
/****************************************/

func Init() {
	agent.AddLifeCycleListener(&PacketStore_LifeCycleListener{})
}

func GetStore(protocol string, agentId int) *PacketStore {
	return packetProtocolMap[protocol].stores[agentId]
}

func GetStores(agentId int) []*PacketStore {
	storeList := make([]*PacketStore, 0)
	for _, ifo := range packetProtocolMap {
		storeList = append(storeList, ifo.stores[agentId])
	}
	return storeList
}

func RegisterPacketProtocol(
	protocol string,
	factory protocol.PacketFactory) {
	if _, contains := packetProtocolMap[protocol]; !contains {
		packetProtocolMap[protocol] = NewPacketProtocolInfo(protocol, factory)
	}
}
