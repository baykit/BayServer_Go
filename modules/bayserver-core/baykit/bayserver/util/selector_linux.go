//go:build linux

package util

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"sync"
)

const OP_READ = 1
const OP_WRITE = 2

type Selector struct {
	sockets map[int]int
	kq      int
	lock    sync.Mutex
}

func NewSelector() *Selector {
	s := Selector{}
	s.sockets = map[int]int{}
	return &s
}

func (s *Selector) Register(skt int, op int) exception.IOException {
	if op&OP_READ != 0 {
		return s.registerRead(skt, s.sockets, true, true)

	} else if op&OP_WRITE != 0 {
		return s.registerWrite(skt, s.sockets, true, true)
	}
	return nil
}

func (s *Selector) Unregister(skt int) exception.IOException {
	err := s.unregisterRead(skt, s.sockets, true, true)
	if err != nil {
		return err
	}

	err = s.unregisterWrite(skt, s.sockets, true, true)
	if err != nil {
		return err
	}

	return nil
}

func (s *Selector) Modify(skt int, op int) exception.IOException {
	var err exception.IOException
	if op&OP_READ != 0 {
		err = s.registerRead(skt, s.sockets, true, true)
	} else {
		err = s.unregisterRead(skt, s.sockets, true, true)
	}
	if err != nil {
		return err
	}

	if op&OP_WRITE != 0 {
		err = s.registerWrite(skt, s.sockets, true, true)
	} else {
		err = s.unregisterWrite(skt, s.sockets, true, true)
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Selector) GetOp(skt int) int {
	op, exists := s.sockets[skt]
	if !exists {
		return -1
	} else {
		return op
	}
}

func (s *Selector) Select(timeoutSec int) (map[int]int, exception.IOException) {
	return nil, nil
}

/****************************************/
/* private functions                    */
/****************************************/

func (s *Selector) registerRead(skt int, sockets map[int]int, needLock bool, addEvent bool) exception.IOException {
	if needLock {
		s.lock.Lock()
	}
	if val, exists := sockets[skt]; exists {
		sockets[skt] = val | OP_READ
	} else {
		sockets[skt] = OP_READ
	}
	if needLock {
		s.lock.Unlock()
	}
	return nil
}

func (s *Selector) registerWrite(skt int, sockets map[int]int, needLock bool, addEvent bool) exception.IOException {
	if needLock {
		s.lock.Lock()
	}
	if val, exists := sockets[skt]; exists {
		sockets[skt] = val | OP_WRITE
	} else {
		sockets[skt] = OP_WRITE
	}
	if needLock {
		s.lock.Unlock()
	}
	return nil
}

func (s *Selector) unregisterRead(skt int, sockets map[int]int, needLock bool, addEvent bool) exception.IOException {
	if needLock {
		s.lock.Lock()
	}
	if val, exists := sockets[skt]; exists {
		if val == OP_READ {
			delete(sockets, skt)
		} else {
			sockets[skt] = OP_WRITE
		}
	}
	if needLock {
		s.lock.Unlock()
	}
	return nil
}

func (s *Selector) unregisterWrite(skt int, sockets map[int]int, needLock bool, addEvent bool) exception.IOException {
	if needLock {
		s.lock.Lock()
	}
	if val, exists := sockets[skt]; exists {
		if val == OP_WRITE {
			delete(sockets, skt)
		} else {
			sockets[skt] = OP_READ
		}
	}
	if needLock {
		s.lock.Unlock()
	}
	return nil
}

func (s *Selector) registerEvent(skt int, filter int16, flags uint16) exception.IOException {
	return nil
}
