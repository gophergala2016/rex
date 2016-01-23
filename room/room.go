// Package room provides a framework for REx servers and clients to communicate
// using arbitrary messages.
package room

import (
	"container/list"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
)

// Config defines the service used for discovery.
type Config struct {
	Service string
}

// TODO: an in-memory list of events imposes many design constraints on
// applications.  Ideally there would be some interface that allows persisting
// and compacting the event list.
type roomBus struct {
	tcp       *net.TCPListener
	hsrv      *http.Server
	eventsin  chan Event
	events    []Event // The history of events
	eventsrdy *sync.Cond
	msgsout   chan Msg
	qm        *list.List
}

// newRoomBus creates a new bus for the server application to use.
//		bus := newRoomBus(":0")
//		msrv := newService(bus, config)
//		// ...
func newRoomBus(addr string, fn func(msg Msg)) (*roomBus, error) {
	b := &roomBus{}

	laddr := &net.TCPAddr{}
	host, _port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	laddr.Port, err = strconv.Atoi(_port)
	if err != nil {
		return nil, fmt.Errorf("invalid address port")
	}
	laddr.IP = net.ParseIP(host)
	if laddr.IP == nil && host != "" {
		return nil, fmt.Errorf("invalid address host")
	}

	b.eventsin = make(chan Event)
	b.eventsrdy = sync.NewCond(new(sync.Mutex))

	b.msgsout = make(chan Msg)
	b.qm = list.New()

	b.tcp, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		return nil, err
	}
	b.hsrv = &http.Server{
		Addr:         addr,
		Handler:      nil, // FIXME
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 0,
	}
	go b.hsrv.Serve(b.tcp)
	go b.dispatch(fn)

	return b, nil
}

// Event broadcasts an event to all client sessions.
func (b *roomBus) Event(c Content) error {
	event := newEvent(0, c, dt.Now)
	b.eventsin <- event
	return nil
}

// dispatch loops forever and calls the Msg handler function in a loop whenever
// there messages.
func (b *roomBus) dispatch(fn func(msg Msg)) {
	for msg := range b.msgsout {
		fn(msg)
	}
}

// Addr returns the string address the bus is listening on for HTTP requests.
func (b *roomBus) Addr() string {
	return b.tcp.Addr().String()
}

type server struct {
	bus  *roomBus
	msrv *mdns.Server
}
