package h2

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common/exception"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type HeaderBlockAnalyzer struct {
	Name   string
	Value  string
	Method string
	Path   string
	Scheme string
	Status string
}

func NewHeaderBlockAnalyzer() *HeaderBlockAnalyzer {
	return &HeaderBlockAnalyzer{}
}

func (a *HeaderBlockAnalyzer) Clear() {
	a.Name = ""
	a.Value = ""
	a.Method = ""
	a.Path = ""
	a.Scheme = ""
	a.Status = ""
}

func (a *HeaderBlockAnalyzer) AnalyzeHeaderBlock(blk *HeaderBlock, tbl *HeaderTable) exception.ProtocolException {
	a.Clear()
	switch blk.op {
	case HEADER_OP_INDEX:
		kv := tbl.Get(blk.index)
		if kv == nil {
			return exception.NewProtocolException("Invalid header index: %d", blk.index)

		}
		a.Name = kv.Name
		a.Value = kv.Value

	case HEADER_OP_KNOWN_HEADER, HEADER_OP_OVERLOAD_KNOWN_HEADER:
		kv := tbl.Get(blk.index)
		if kv == nil {
			return exception.NewProtocolException("Invalid header index: %d", blk.index)
		}
		a.Name = kv.Name
		a.Value = blk.value
		if blk.op == HEADER_OP_OVERLOAD_KNOWN_HEADER {
			tbl.Insert(a.Name, a.Value)
		}

	case HEADER_OP_NEW_HEADER:
		a.Name = blk.name
		a.Value = blk.value
		tbl.Insert(a.Name, a.Value)

	case HEADER_OP_UNKNOWN_HEADER:
		a.Name = blk.name
		a.Value = blk.value

	case HEADER_OP_UPDATE_DYNAMIC_TABLE_SIZE:
		tbl.SetSize(blk.size)

	default:
		bayserver.FatalError(exception2.NewSink("Invalid op: %d", blk.op))
	}

	if a.Name != "" && a.Name[0] == ':' {
		switch a.Name {
		case PSEUDO_HEADER_AUTHORITY:
			a.Name = "host"

		case PSEUDO_HEADER_METHOD:
			a.Method = a.Value

		case PSEUDO_HEADER_PATH:
			a.Path = a.Value

		case PSEUDO_HEADER_SCHEME:
			a.Scheme = a.Value

		case PSEUDO_HEADER_STATUS:
			a.Status = a.Value
		}
	}
	return nil
}
