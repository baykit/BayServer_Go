package objectstore

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/strutil"
)

type ObjectStore struct {
	freeList   []util.Reusable
	activeList []util.Reusable
	factory    ObjectFactory[util.Reusable]
}

func NewObjectStore(fct ObjectFactory[util.Reusable]) *ObjectStore {
	return &ObjectStore{
		freeList:   []util.Reusable{},
		activeList: []util.Reusable{},
		factory:    fct,
	}
}

func (st *ObjectStore) Rent() util.Reusable {
	var obj util.Reusable
	//baylog.debug(owner + " rent freeList=" + freeList);
	if len(st.freeList) == 0 {
		obj = st.factory()

	} else {
		obj = st.freeList[0]
		st.freeList = st.freeList[1:]
	}

	if obj == nil {
		bayserver.FatalError(exception.NewSink("Object is null"))
	}

	st.activeList = append(st.activeList, obj)
	//baylog.debug(owner + " rent object " + obj);
	return obj
}

func (st *ObjectStore) Return(obj util.Reusable, reuse bool) {
	if arrayutil.Contains(st.freeList, obj) {
		bayserver.FatalError(exception.NewSink("This object already returned: %s", obj))
	}

	if !arrayutil.Contains(st.activeList, obj) {
		bayserver.FatalError(exception.NewSink("This object is not active: %s", obj))
	}

	st.activeList, _ = arrayutil.RemoveObject(st.activeList, obj)
	if reuse {
		st.freeList = append(st.freeList, obj)
		obj.Reset()
	}
}

// Print memory usage
func (st *ObjectStore) printUsage(indent int) {
	baylog.Info("%sfree list: %d", strutil.Indent(indent), len(st.freeList))
	baylog.Info("%sactive list: %d", strutil.Indent(indent), len(st.activeList))
	if baylog.IsDebugMode() {
		for _, obj := range st.activeList {
			baylog.Debug("%s%s", strutil.Indent(indent+1), obj)
		}
	}
}

func (st *ObjectStore) PrintUsage(indent int) {
	baylog.Info("%sfree list: %d", strutil.Indent(indent), len(st.freeList))
	baylog.Info("%sactive list: %d", strutil.Indent(indent), len(st.activeList))
	if baylog.IsDebugMode() {
		for _, obj := range st.activeList {
			baylog.Debug("%s%s", strutil.Indent(indent+1), obj)
		}
	}
}
