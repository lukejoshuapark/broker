package broker

import (
	"sync"

	"github.com/lukejoshuapark/infchan"
)

// Broker provides a mpmc "fan-out" broker.  The underlying implementation is
// buffered to ensure isolation between the In channel and each of the
// subscribers.  All methods on `Broker` are thread-safe.
//
// When `Close` is called, the Broker will enter a shutdown state.  In this
// state, sending any new messages will cause the broker to panic.  Any pending
// messages will still be received by downstream consumers, however.  When all
// pending messages have been received, the corresponding channels will close
// automatically.
//
// Do not ever manually close the channels returned by any of the methods on
// Broker.  Calling Close on the broker will ensure a clean shutdown.
type Broker[T any] struct {
	rwm  *sync.RWMutex
	in   infchan.Channel[T]
	outs map[<-chan T]infchan.Channel[T]
}

func NewBroker[T any]() *Broker[T] {
	b := &Broker[T]{
		rwm:  &sync.RWMutex{},
		in:   infchan.NewChannel[T](),
		outs: map[<-chan T]infchan.Channel[T]{},
	}

	go b.process()
	return b
}

func (b *Broker[T]) Close() {
	b.in.Close()
}

func (b *Broker[T]) In() chan<- T {
	return b.in.In()
}

func (b *Broker[T]) Subscribe() <-chan T {
	b.rwm.Lock()
	defer b.rwm.Unlock()

	if b.outs == nil {
		panic("cannot subscribe to a closed broker")
	}

	out := infchan.NewChannel[T]()
	outc := out.Out()

	b.outs[outc] = out
	return outc
}

func (b *Broker[T]) Unsubscribe(c <-chan T) {
	b.rwm.Lock()
	defer b.rwm.Unlock()

	if b.outs == nil {
		return
	}

	out, ok := b.outs[c]
	if !ok {
		return
	}

	out.Close()
	delete(b.outs, c)
}

func (b *Broker[T]) process() {
	for v := range b.in.Out() {
		func() {
			b.rwm.RLock()
			defer b.rwm.RUnlock()

			for _, out := range b.outs {
				out.In() <- v
			}
		}()
	}

	b.rwm.Lock()
	defer b.rwm.Unlock()

	for _, out := range b.outs {
		out.Close()
	}

	b.outs = nil
}
