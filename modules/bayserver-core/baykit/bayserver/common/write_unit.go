package common

import (
	"net"
)

type WriteUnit struct {
	Buf      []byte
	Addr     net.Addr
	Tag      interface{}
	listener DataConsumeListener
}

func NewWriteUnit(buf []byte, adr net.Addr, tag interface{}, lis DataConsumeListener) *WriteUnit {
	return &WriteUnit{
		Buf:      buf,
		Addr:     adr,
		Tag:      tag,
		listener: lis,
	}
}

func (w *WriteUnit) Done() {
	if w.listener != nil {
		w.listener()
	}
}
