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

// Bytes returns a Content containing the b as its data.
func Bytes(b []byte) Content {
	return contentBytes(b)
}

// String returns a Content containing the s as its data.
func String(s string) Content {
	return contentString(s)
}

type contentBytes []byte

var _ Content = contentBytes(nil)

func (c contentBytes) Data() []byte {
	return []byte(c)
}

func (c contentBytes) Text() string {
	return string(c)
}

type contentString string

var _ Content = contentString("")

func (c contentString) Data() []byte {
	return []byte(c)
}

func (c contentString) Text() string {
	return string(c)
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
	event := &simpleEvent{t(), c}
	return event
}

type simpleEvent struct {
	t Time
	Content
}

var _ Event = &simpleEvent{}

func (event *simpleEvent) Time() Time {
	return event.t
}

func newMsg(session string, c Content, t func() Time) Msg {
	msg := &simpleMsg{session, t(), c}
	return msg
}

type simpleMsg struct {
	sess string
	t    Time
	Content
}

var _ Msg = &simpleMsg{}

func (msg *simpleMsg) Session() string {
	return msg.sess
}

func (msg *simpleMsg) Time() Time {
	return msg.t
}
