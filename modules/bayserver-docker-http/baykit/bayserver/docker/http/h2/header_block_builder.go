package h2

import (
	exception2 "bayserver-core/baykit/bayserver/common/exception"
)

type HeaderBlockBuilder struct {
}

func NewHeaderBlockBuilder() *HeaderBlockBuilder {
	return &HeaderBlockBuilder{}
}

func (b *HeaderBlockBuilder) BuildHeaderBlock(name string, value string, tbl *HeaderTable) (*HeaderBlock, exception2.ProtocolException) {
	idxList := tbl.GetByName(name)

	var blk *HeaderBlock = nil
	for _, idx := range idxList {
		kv := tbl.Get(idx)
		if kv == nil {
			return nil, exception2.NewProtocolException("Invalid header index: %d", blk.index)
		}
		if kv != nil && value == kv.Value {
			blk = NewHeaderBlock()
			blk.op = HEADER_OP_INDEX
			blk.index = idx
			break
		}
	}
	if blk == nil {
		blk = NewHeaderBlock()
		if len(idxList) > 0 {
			blk.op = HEADER_OP_KNOWN_HEADER
			blk.index = idxList[0]
			blk.value = value
		} else {
			blk.op = HEADER_OP_UNKNOWN_HEADER
			blk.name = name
			blk.value = value
		}
	}

	return blk, nil

}
