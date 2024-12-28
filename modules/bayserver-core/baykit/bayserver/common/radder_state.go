package common

import (
	"bayserver-core/baykit/bayserver/rudder"
	"sync"
	"time"
)

type RudderState struct {
	Rudder      rudder.Rudder
	Transporter Transporter
	Multiplexer Multiplexer

	LastAccessTime int64
	Connecting     bool
	Closing        bool
	ReadBuf        []byte
	ReadPos        int
	WriteQueue     []*WriteUnit
	QueueLock      sync.Mutex
	ReadLock       sync.Mutex
	WriteLock      sync.Mutex
	CloseLock      sync.Mutex
	Reading        bool
	Writing        bool
	BytesRead      int
	BytesWrote     int
	Closed         bool
	Finale         bool
}

func NewRudderState(rd rudder.Rudder, tp Transporter) *RudderState {
	return &RudderState{
		Rudder:         rd,
		Transporter:    tp,
		LastAccessTime: 0,
		Connecting:     false,
		Closing:        false,
		ReadBuf:        make([]byte, 8192),
		ReadPos:        0,
		WriteQueue:     make([]*WriteUnit, 0),
		Reading:        false,
		Writing:        false,
		BytesRead:      0,
		BytesWrote:     0,
		Closed:         false,
		Finale:         false,
	}
}

func (st *RudderState) String() string {
	return "RudderState"
}

func (st *RudderState) Access() {
	st.LastAccessTime = time.Now().Unix()
}

func (st *RudderState) End() {
	st.Finale = true
}
