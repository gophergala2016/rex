package rexdemo

import (
	"sync"
	"time"

	"github.com/gophergala2016/rex/room"
)

// Room is the room used by clients and servers for the demo.
var Room = &room.Room{
	Name:    "REx Demo",
	Service: "_rexdemo._tcp.",
}

// Demo is the state of a demo a copy of the state is present in the server and
// all clients.
type Demo struct {
	Mut     *sync.Mutex `json:"-"`
	Counter int         `json:"counter"`
	Last    time.Time   `json:"last"`
}

// NewDemo returns a new Demo object
func NewDemo() *Demo {
	// It's a little weird that a pointer is preferred over sync.Mutex.  But
	// due to how state is shored be the client and server a reference makes
	// more sense.
	return &Demo{
		Mut: new(sync.Mutex),
	}
}
