package warpshipstore

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common/objectstore"
	"bayserver-core/baykit/bayserver/common/warpship"
	"bayserver-core/baykit/bayserver/common/warpship/impl"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/strutil"
	"sync"
)

type WarpShipStore struct {
	objectstore.ObjectStore
	keepList []warpship.WarpShip
	busyList []warpship.WarpShip
	maxShips int
	lock     sync.Mutex
}

func NewWarpShipStore(maxShips int) *WarpShipStore {
	store := WarpShipStore{
		ObjectStore: *objectstore.NewObjectStore(func() util.Reusable { return impl.NewWarpShip() }),
		keepList:    make([]warpship.WarpShip, 0),
		busyList:    make([]warpship.WarpShip, 0),
		maxShips:    maxShips,
	}
	return &store
}

/****************************************/
/*  Public functions                    */
/****************************************/

func (st *WarpShipStore) Rent() warpship.WarpShip {
	st.lock.Lock()
	defer st.lock.Unlock()

	if st.maxShips > 0 && st.count() >= st.maxShips {
		return nil
	}

	var wsip warpship.WarpShip = nil
	if len(st.keepList) == 0 {
		wsip = st.ObjectStore.Rent().(warpship.WarpShip)
		if wsip == nil {
			return nil
		}

	} else {
		wsip = st.keepList[len(st.keepList)-1]
		st.keepList = arrayutil.RemoveAt(st.keepList, len(st.keepList)-1)
	}

	if wsip == nil {
		bayserver.FatalError(exception.NewSink("BUG! warp ship is null"))
	}

	st.busyList = append(st.busyList, wsip)

	return wsip
}

/**
 * Keep ship which connection is alive
 */

func (st *WarpShipStore) Keep(wsip warpship.WarpShip) {
	st.lock.Lock()
	defer st.lock.Unlock()

	var removed bool
	st.busyList, removed = arrayutil.RemoveObject(st.busyList, wsip)
	if !removed {
		baylog.Error("BUG: %s not in busy list", wsip)
	}
	st.keepList = append(st.keepList, wsip)
}

/**
 * Return ship which connection is closed
 */

func (st *WarpShipStore) Return(wsip warpship.WarpShip) {
	st.lock.Lock()
	defer st.lock.Unlock()

	var removedFromKeep, removedFromBusy bool
	st.keepList, removedFromKeep = arrayutil.RemoveObject(st.keepList, wsip)
	st.busyList, removedFromBusy = arrayutil.RemoveObject(st.busyList, wsip)
	if !removedFromKeep && !removedFromBusy {
		baylog.Error("BUB: %s not in both keep and busy list")
	}

	st.ObjectStore.Return(wsip, true)
}

func (st *WarpShipStore) PrintUsage(indent int) {
	baylog.Info("%sWarpShipStore Usage:", strutil.Indent(indent))
	baylog.Info("%skeepList: %d", strutil.Indent(indent+1), len(st.keepList))
	if baylog.IsDebugMode() {
		for _, obj := range st.keepList {
			baylog.Debug("%s%s", strutil.Indent(indent+1), obj)
		}
	}
	baylog.Info("%sbusyList: %d", strutil.Indent(indent+1), len(st.busyList))
	if baylog.IsDebugMode() {
		for _, obj := range st.busyList {
			baylog.Debug("%s%s", strutil.Indent(indent+1), obj)
		}
	}
	st.ObjectStore.PrintUsage(indent)
}

func (st *WarpShipStore) count() int {
	return len(st.keepList) + len(st.busyList)
}
