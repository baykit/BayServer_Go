package multiplexer

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/util/exception"
	"time"
)

type PipeRecipient struct {
	common.Recipient

	pipe   chan int
	ticker *time.Ticker
}

func NewPipeRecipient() common.Recipient {
	p := PipeRecipient{
		pipe:   make(chan int, 10),
		ticker: time.NewTicker(10 * time.Second),
	}

	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		for {
			<-p.ticker.C
			p.Wakeup()
		}
	}()
	return &p
}

func (p *PipeRecipient) Receive(wait bool) (bool, exception.IOException) {

	if wait {
		_, ok := <-p.pipe
		if !ok {
			return false, exception.NewIOException("Channel closed")

		} else {
			return true, nil
		}

	} else {
		select {
		case _, ok := <-p.pipe:
			if !ok {
				return false, exception.NewIOException("Channel closed")

			} else {
				return true, nil
			}

		default:
			return false, nil
		}
	}
}

func (p *PipeRecipient) Wakeup() {
	p.pipe <- 0
}

func (p *PipeRecipient) End() {
	p.ticker.Stop()
}
