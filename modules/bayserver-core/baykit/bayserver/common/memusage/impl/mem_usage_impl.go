package impl

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/inboundship/inboundshipstore"
	"bayserver-core/baykit/bayserver/common/memusage"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-core/baykit/bayserver/tour/tourstore"
	"bayserver-core/baykit/bayserver/util/baylog"
)

/** Agent ID => MemUsage */
var memUsages = map[int]memusage.MemUsage{}

/****************************************/
/*  Type MemUsage_LifeCycleListener     */
/****************************************/

type MemUsage_LifeCycleListener struct {
	common.LifecycleListener
}

func NewMemUsage_LifeCycleListener() *MemUsage_LifeCycleListener {
	return &MemUsage_LifeCycleListener{}
}

func (l *MemUsage_LifeCycleListener) Add(agentId int) {
	memUsages[agentId] = NewMemUsage(agentId)
}

func (l *MemUsage_LifeCycleListener) Remove(agentId int) {
	delete(memUsages, agentId)
}

/****************************************/
/*  Type MemUsage                       */
/****************************************/

type MemUsageImpl struct {
	AgentId int
}

func NewMemUsage(agtId int) memusage.MemUsage {
	return &MemUsageImpl{AgentId: agtId}
}

func (m *MemUsageImpl) PrintUsage(indent int) {
	baylog.Info("Agent#%d MemUsage", m.AgentId)
	inboundshipstore.Get(m.AgentId).PrintUsage(indent + 1)
	for _, store := range protocolhandlerstore.GetStores(m.AgentId) {
		store.PrintUsage(indent + 1)
	}
	for _, store := range packetstore.GetStores(m.AgentId) {
		store.PrintUsage(indent + 1)
	}
	tourstore.GetStore(m.AgentId).PrintUsage(indent + 1)
	for _, city := range bayserver.Cities().Cities() {
		m.PrintCityUsage(nil, city, indent)
	}
	for _, port := range bayserver.Ports() {
		for _, city := range port.Cities() {
			m.PrintCityUsage(port, city, indent)
		}
	}
}

func (m *MemUsageImpl) PrintCityUsage(port docker.Port, city docker.City, indent int) {
	/*
	   pname := ""
	   if port != nil {
	       pname = "@" + fmt.Sprint(port)
	   }
	   for _, club := range city.Clubs() {
	       if wb, ok := club.(WarpBase), ok {
	           baylog.Info("%sClub(%s%s) Usage:", strutil.Indent(indent), club, pname);
	           wb.GetShipStore(m.AgentId).PrintUsage(indent+1);
	       }
	   });
	   city.towns().forEach(town -> {
	       town.clubs().forEach(club -> {
	           if (club instanceof WarpBase) {
	               BayLog.info("%sClub(%s%s) Usage:", StringUtil.indent(indent), club, pname);
	               ((WarpBase) club).getShipStore(agentId).printUsage(indent+1);
	           }
	       });
	   });
	*/
}

/****************************************/
/*  Static methods                      */
/****************************************/

func Init() {
	agent.AddLifeCycleListener(NewMemUsage_LifeCycleListener())
	memusage.NewMemUsage = NewMemUsage
	memusage.Get = Get
}

func Get(agtId int) memusage.MemUsage {
	return memUsages[agtId]
}
