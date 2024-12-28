package inboundshipstore

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/common"
	inboundship "bayserver-core/baykit/bayserver/common/inboundship/impl"
	"bayserver-core/baykit/bayserver/common/objectstore"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/strutil"
)

var inboundShipStores = map[int]*InboundShipStore{}

type InboundShipStore_LifeCycleListener struct {
	common.LifecycleListener
}

type InboundShipStore struct {
	objectstore.ObjectStore
}

func (l *InboundShipStore_LifeCycleListener) Add(agentId int) {
	inboundShipStores[agentId] = NewInboundShipStore()
}

func (l *InboundShipStore_LifeCycleListener) Remove(agentId int) {
	delete(inboundShipStores, agentId)
}

func NewInboundShipStore() *InboundShipStore {
	return &InboundShipStore{
		ObjectStore: *objectstore.NewObjectStore(func() util.Reusable { return inboundship.NewInboundShip() }),
	}
}

func (s *InboundShipStore) PrintUsage(indent int) {
	baylog.Info("%sInboundShipStore Usage:", strutil.Indent(indent))
	s.ObjectStore.PrintUsage(indent + 1)
}

func Init() {
	agent.AddLifeCycleListener(&InboundShipStore_LifeCycleListener{})
}

func Get(agentId int) *InboundShipStore {
	return inboundShipStores[agentId]
}
