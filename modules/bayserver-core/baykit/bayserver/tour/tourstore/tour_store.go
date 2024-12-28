package tourstore

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/strutil"
	"sync"
)

const MAX_TOURS = 128

var stores = map[int]*TourStore{}
var maxCount = 0

/************************************************/
/* Type TourStore_LifeCycleListener             */
/************************************************/

type TourStore_LifeCycleListener struct {
	common.LifecycleListener
}

func (l *TourStore_LifeCycleListener) Add(agentId int) {
	stores[agentId] = NewTourStore()
}

func (*TourStore_LifeCycleListener) Remove(agentId int) {
	delete(stores, agentId)
}

type TourStore struct {
	freeTours     []tour.Tour
	activeTourMap map[int64]tour.Tour
	lock          sync.Mutex
}

func NewTourStore() *TourStore {
	return &TourStore{
		freeTours:     []tour.Tour{},
		activeTourMap: make(map[int64]tour.Tour),
	}
}

/************************************************/
/* Public functions                             */
/************************************************/

func (st *TourStore) Get(key int64) tour.Tour {
	return st.activeTourMap[key]
}

func (st *TourStore) Rent(key int64, force bool) tour.Tour {
	tur := st.Get(key)
	if tur != nil {
		bayserver.FatalError(exception.NewSink("Tour is active: %s", tur))
	}

	if len(st.freeTours) > 0 {
		tur = st.freeTours[0]
		st.freeTours = st.freeTours[1:]

	} else {
		if !force && len(st.activeTourMap) >= maxCount {
			return nil

		} else {
			tur = impl.NewTour()
		}
	}

	st.activeTourMap[key] = tur
	return tur
}

func (st *TourStore) Return(key int64) {
	st.lock.Lock()
	defer st.lock.Unlock()

	if _, exists := st.activeTourMap[key]; !exists {
		bayserver.FatalError(exception.NewSink("Tour is not active key=: %d", key))
	}

	//baylog.Debug("Return tour: key=%d Active tour count: before=%d", key, len(st.activeTourMap))

	tur := st.activeTourMap[key]
	delete(st.activeTourMap, key)

	tur.Reset()
	st.freeTours = append(st.freeTours, tur)
}

func (st *TourStore) PrintUsage(indent int) {
	baylog.Info("%sTour store usage:", strutil.Indent(indent))
	baylog.Info("%sfreeList: %d", strutil.Indent(indent+1), len(st.freeTours))
	baylog.Info("%sactiveList: %d", strutil.Indent(indent+1), len(st.activeTourMap))
	if baylog.IsDebugMode() {
		for _, tur := range st.activeTourMap {
			baylog.Debug("%s%s", strutil.Indent(indent+2), tur)
		}
	}
}

func Init(maxTourCount int) {
	maxCount = maxTourCount
	agent.AddLifeCycleListener(&TourStore_LifeCycleListener{})
}

func GetStore(agentId int) *TourStore {
	return stores[agentId]
}
