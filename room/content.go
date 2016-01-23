package room

// Content is application data that is transmitted over a bus.
type Content interface {
	// Data returns the event content is a slice of bytes (even if the content
	// is stored as a string).  Applications should use whichever of Data or
	// Text methods they need and any translation necessary will happen as fast
	// as possible.
	Data() []byte

	// Text is like Data but returns a string representation of event data and
	// may be optmized to do that.
	Text() string
}

// Event is a broadcast message from the server to all clients.  Unlike Msg an
// event does not have an associated session identifier because it is intended
// for all clients.
type Event interface {
	Time() Time

	Content
}

// Msg is a piece of data originating from a client application.
type Msg interface {
	// Session returns an identifier for the client (session) that originated
	// the message.
	Session() string

	Time() Time

	Content
}

func newEvent(c Content, t func() Time) Event {
	return nil
}

func newMsg(session string, c Content, t func() Time) Msg {
	return nil
}
