package h2

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
 * HPACK spec
 *
 * https://datatracker.ietf.org/doc/html/rfc7541
 */

const PSEUDO_HEADER_AUTHORITY = ":authority"

const PSEUDO_HEADER_METHOD = ":method"

const PSEUDO_HEADER_PATH = ":path"

const PSEUDO_HEADER_SCHEME = ":scheme"

const PSEUDO_HEADER_STATUS = ":status"

var staticTable = newHeaderTable()
var staticSize = 0

type HeaderTable struct {
	idxMap   []*util.KeyVal
	addCount int
	nameMap  map[string][]int
}

func newHeaderTable() *HeaderTable {
	h := &HeaderTable{
		idxMap:   make([]*util.KeyVal, 0),
		addCount: 0,
		nameMap:  make(map[string][]int),
	}
	return h
}

func (t *HeaderTable) Get(idx int) *util.KeyVal {
	if idx <= 0 || idx > staticSize+len(t.idxMap) {
		baylog.Error("Invalid Index: idx=%d static=%d dynamic=%d", idx, staticSize, len(t.idxMap))
		return nil
	}

	var kv *util.KeyVal
	if idx <= staticSize {
		kv = staticTable.idxMap[idx-1]

	} else {
		kv = t.idxMap[(idx-staticSize)-1]

	}
	return kv
}

func (t *HeaderTable) GetByName(name string) []int {
	dynamicList := t.nameMap[name]
	staticList := staticTable.nameMap[name]

	idxList := make([]int, 0)
	if staticList != nil {
		idxList = append(idxList, staticList...)
	}
	if dynamicList != nil {
		for _, idx := range dynamicList {
			readIndex := t.addCount - idx + staticSize
			idxList = append(idxList, readIndex)
		}
	}
	return idxList
}

func (t *HeaderTable) Insert(name string, value string) {
	t.idxMap = append([]*util.KeyVal{util.NewKeyVal(name, value)}, t.idxMap...)
	t.addCount++
	t.addToNameMap(name, t.addCount)
}

func (t *HeaderTable) put(idx int, name string, value string) {
	if idx != len(t.idxMap)+1 {
		bayserver.FatalError(exception.NewSink("Illegal State"))
	}
	t.idxMap = append(t.idxMap, util.NewKeyVal(name, value))
	t.addToNameMap(name, idx)
}

func (t *HeaderTable) addToNameMap(name string, idx int) {
	idxList := t.nameMap[name]
	if idxList == nil {
		idxList = make([]int, 0)
	}
	idxList = append(idxList, idx)
	t.nameMap[name] = idxList
}

func (t *HeaderTable) SetSize(size int) {

}

func CreateDynamicTable() *HeaderTable {
	t := newHeaderTable()
	return t
}

func HeaderTableInit() bool {
	staticTable.put(1, PSEUDO_HEADER_AUTHORITY, "")
	staticTable.put(2, PSEUDO_HEADER_METHOD, "GET")
	staticTable.put(3, PSEUDO_HEADER_METHOD, "POST")
	staticTable.put(4, PSEUDO_HEADER_PATH, "/")
	staticTable.put(5, PSEUDO_HEADER_PATH, "/index.html")
	staticTable.put(6, PSEUDO_HEADER_SCHEME, "http")
	staticTable.put(7, PSEUDO_HEADER_SCHEME, "https")
	staticTable.put(8, PSEUDO_HEADER_STATUS, "200")
	staticTable.put(9, PSEUDO_HEADER_STATUS, "204")
	staticTable.put(10, PSEUDO_HEADER_STATUS, "206")
	staticTable.put(11, PSEUDO_HEADER_STATUS, "304")
	staticTable.put(12, PSEUDO_HEADER_STATUS, "400")
	staticTable.put(13, PSEUDO_HEADER_STATUS, "404")
	staticTable.put(14, PSEUDO_HEADER_STATUS, "500")
	staticTable.put(15, "accept-charset", "")
	staticTable.put(16, "accept-encoding", "gzip, deflate")
	staticTable.put(17, "accept-language", "")
	staticTable.put(18, "accept-ranges", "")
	staticTable.put(19, "accept", "")
	staticTable.put(20, "access-control-allow-origin", "")
	staticTable.put(21, "age", "")
	staticTable.put(22, "allow", "")
	staticTable.put(23, "authorization", "")
	staticTable.put(24, "cache-control", "")
	staticTable.put(25, "content-disposition", "")
	staticTable.put(26, "content-encoding", "")
	staticTable.put(27, "content-language", "")
	staticTable.put(28, "content-length", "")
	staticTable.put(29, "content-location", "")
	staticTable.put(30, "content-range", "")
	staticTable.put(31, "content-type", "")
	staticTable.put(32, "cookie", "")
	staticTable.put(33, "date", "")
	staticTable.put(34, "etag", "")
	staticTable.put(35, "expect", "")
	staticTable.put(36, "expires", "")
	staticTable.put(37, "from", "")
	staticTable.put(38, "host", "")
	staticTable.put(39, "if-match", "")
	staticTable.put(40, "if-modified-since", "")
	staticTable.put(41, "if-none-match", "")
	staticTable.put(42, "if-range", "")
	staticTable.put(43, "if-unmodified-since", "")
	staticTable.put(44, "last-modified", "")
	staticTable.put(45, "link", "")
	staticTable.put(46, "location", "")
	staticTable.put(47, "max-forwards", "")
	staticTable.put(48, "proxy-authenticate", "")
	staticTable.put(49, "proxy-authorization", "")
	staticTable.put(50, "range", "")
	staticTable.put(51, "referer", "")
	staticTable.put(52, "refresh", "")
	staticTable.put(53, "retry-after", "")
	staticTable.put(54, "server", "")
	staticTable.put(55, "set-cookie", "")
	staticTable.put(56, "strict-transport-security", "")
	staticTable.put(57, "transfer-encoding", "")
	staticTable.put(58, "user-agent", "")
	staticTable.put(59, "vary", "")
	staticTable.put(60, "via", "")
	staticTable.put(61, "www-authenticate", "")

	staticSize = len(staticTable.idxMap)
	return true
}

var dummy = HeaderTableInit()
