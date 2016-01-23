package room

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"
)

// Handler is a bus message handler.
type Handler interface {
	HandleMessage(ctx context.Context, msg Msg)
}

// handlerFunc implement Handler
type handlerFunc func(context.Context, Msg)

func hfunc(fn func(ctx context.Context, msg Msg)) Handler {
	return handlerFunc(fn)
}

func (fn handlerFunc) HandleMessage(ctx context.Context, msg Msg) {
	fn(ctx, msg)
}

// Bus is the communication bus for a Server.
type Bus struct {
	ctx  context.Context
	term chan struct{}

	hmut     sync.RWMutex
	handlers []Handler

	sub       chan chan<- int
	eventsin  chan Event
	events    []Event // The history of events
	eventsrdy *sync.Cond
	msgs      chan Msg
}

// NewBus initializes and returns a new Bus.
func NewBus(ctx context.Context, handlers ...Handler) *Bus {
	b := &Bus{}
	b.init()
	b.handlers = handlers
	go b.msgLoop()
	go b.eventLoop()
	return b
}

func (b *Bus) close() {
	close(b.term)
}

func (b *Bus) init() {
	b.term = make(chan struct{})
	b.sub = make(chan chan<- int)
	b.eventsin = make(chan Event)
	b.eventsrdy = sync.NewCond(&sync.Mutex{})
	b.msgs = make(chan Msg)
}

// Event broadcasts an event to all Subscription.
func (b *Bus) Event(c Content) error {
	event := newEvent(0, c, dt.Now)
	b.eventsin <- event
	return nil
}

// Message is called by a subscriber to signal back to the bus owner via
// b.handler.
func (b *Bus) Message(session string, c Content) error {
	msg := newMsg(session, c, dt.Now)
	b.msgs <- msg
	return nil
}

// AddHandler changes the bus message handler.
func (b *Bus) AddHandler(h Handler) {
	b.hmut.Lock()
	defer b.hmut.Unlock()
	b.handlers = append(b.handlers, h)
}

func (b *Bus) handle(msg Msg) {
	b.hmut.RLock()
	defer b.hmut.RUnlock()
	ctx := withBus(b.ctx, b)
	for _, h := range b.handlers {
		h.HandleMessage(ctx, msg)
	}
}

// msgLoop dispatches messages passed in with b.Message to b.handler.  Calls to
// b.handler as serialized.  Concurrency must be handled at a higher level of
// abstraction.
func (b *Bus) msgLoop() {
	for {
		select {
		case <-b.term:
			// FIXME notify future callers of b.Message()
			return
		case msg := <-b.msgs:
			b.handle(msg)
		}
	}
}

func (b *Bus) eventLoop() {
	defer b.eventsrdy.Broadcast()

	for {
		select {
		case <-b.term:
			return
		case event := <-b.eventsin:
			//log.Printf("event! %v", event.Text())
			i := uint64(len(b.events))
			ievent := newEvent(i, event, event.Time)
			b.eventsrdy.L.Lock()
			//log.Printf("locked!")
			b.events = append(b.events, ievent)
			b.eventsrdy.Broadcast()
			b.eventsrdy.L.Unlock()
			//log.Printf("unlocked!")
		}
	}
}

// Subscribe returns a new Subscription that new events from b.
func (b *Bus) Subscribe(start int) *Subscription {
	s := &Subscription{
		term: make(chan struct{}),
		req:  make(chan chan<- Event),
	}
	go b.fulfill(start, s)
	return s
}

func (b *Bus) fulfill(start int, s *Subscription) {
	defer close(s.term)

	i := start
	for {
		b.eventsrdy.L.Lock()
		events := b.events
		for i >= len(events) {
			select {
			case <-b.term:
				return
			default:
			}
			//log.Printf("events: %d", len(events))
			b.eventsrdy.Wait()
			events = b.events
		}
		b.eventsrdy.L.Unlock()
		//log.Printf("here we go: %d", len(events))
		for _, event := range events[i:] {
			select {
			case <-b.term:
				return
			case c, ok := <-s.req:
				if !ok {
					return
				}
				c <- event
				i++
			}
		}
	}
}

// Unsubscribe removes s from the recipients of b's events.  After Unsubscribe
// returns no further events will be received in calls to s.Next().
func (b *Bus) Unsubscribe(s *Subscription) {
	s.close()
}

// Subscription represents a remote client that needs to receive messages from
// a Bus.
type Subscription struct {
	term  chan struct{}
	req   chan chan<- Event
	event Event
}

// Event returns the last received Event.
func (s *Subscription) Event() Event {
	return s.event
}

// Next waits for the next event to be received over the channel and returns
// it.  If the subscription is terminated, or values is received over timeout
// Next will return a value value.  Otherwise Next returns true and Event will
// return the event received.
func (s *Subscription) Next(timeout <-chan time.Time) (ok bool) {
	c := make(chan Event)
	select {
	case <-timeout:
		return false
	case <-s.term:
		//log.Printf("sub: terminated")
		return false
	case s.req <- c:
		//log.Printf("next: 1")
		s.event, ok = <-c
		//log.Printf("next: 2 %v", ok)
		return ok
	}
}

func (s *Subscription) close() {
	close(s.req)
}

// Broadcast sends a broadcast event to all clients connected to the Bus
// associated with ctx.
func Broadcast(ctx context.Context, content Content) error {
	b := contextBus(ctx)
	if b == nil {
		return fmt.Errorf("context has no associated bus")
	}
	return b.Event(content)
}

type busContextKey struct{}

func withBus(ctx context.Context, b *Bus) context.Context {
	return context.WithValue(ctx, busContextKey{}, b)
}

func contextBus(ctx context.Context) *Bus {
	b, _ := ctx.Value(busContextKey{}).(*Bus)
	return b
}
