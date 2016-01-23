package room

import "sync"

type bus struct {
	handler   func(msg Msg)
	term      chan struct{}
	sub       chan chan<- int
	eventsin  chan Event
	events    []Event // The history of events
	eventsrdy *sync.Cond
	msgs      chan Msg
}

func newBus(handler func(msg Msg)) *bus {
	b := &bus{}
	b.init()
	b.handler = handler
	go b.msgLoop()
	go b.eventLoop()
	return b
}

func (b *bus) close() {
	close(b.term)
}

func (b *bus) init() {
	b.term = make(chan struct{})
	b.sub = make(chan chan<- int)
	b.eventsin = make(chan Event)
	b.eventsrdy = sync.NewCond(&sync.Mutex{})
	b.msgs = make(chan Msg)
}

// Event broadcasts an event to all subscriptions.
func (b *bus) Event(c Content) error {
	event := newEvent(0, c, dt.Now)
	b.eventsin <- event
	return nil
}

// Message is called by a subscriber to signal back to the bus owner via
// b.handler.
func (b *bus) Message(session string, c Content) error {
	msg := newMsg(session, c, dt.Now)
	b.msgs <- msg
	return nil
}

// msgLoop dispatches messages passed in with b.Message to b.handler.  Calls to
// b.handler as serialized.  Concurrency must be handled at a higher level of
// abstraction.
func (b *bus) msgLoop() {
	for {
		select {
		case <-b.term:
			// FIXME notify future callers of b.Message()
			return
		case msg := <-b.msgs:
			if b.handler != nil {
				b.handler(msg)
			}
		}
	}
}

func (b *bus) eventLoop() {
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

func (b *bus) Subscribe(start int) *subscription {
	s := &subscription{
		term: make(chan struct{}),
		req:  make(chan chan<- Event),
	}
	go b.fulfill(start, s)
	return s
}

func (b *bus) fulfill(start int, s *subscription) {
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

func (b *bus) Unsubscribe(s *subscription) {
	s.close()
}

type subscription struct {
	term  chan struct{}
	req   chan chan<- Event
	event Event
}

func (s *subscription) Event() Event {
	return s.event
}

func (s *subscription) Next() (ok bool) {
	c := make(chan Event)
	select {
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

func (s *subscription) close() {
	close(s.req)
}
