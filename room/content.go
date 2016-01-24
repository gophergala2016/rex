package room

import "encoding/json"

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
	Index() uint64

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

func newEvent(i uint64, c Content, t func() Time) Event {
	event := &simpleEvent{i, t(), c}
	return event
}

type jsonEvent struct {
	I     uint64 `json:"index"`
	T     Time   `json:"time"`
	D     string `json:"data"`
	Event `json:"-"`
}

func newJSONEvent(event Event) *jsonEvent {
	if event == nil {
		return &jsonEvent{}
	}
	return &jsonEvent{
		I: event.Index(),
		T: event.Time(),
		D: event.Text(),
	}
}

func (event *jsonEvent) MarshalJSON() ([]byte, error) {
	type E jsonEvent
	return json.Marshal((*E)(event))
}

func (event *jsonEvent) UnmarshalJSON(b []byte) error {
	type E jsonEvent
	err := json.Unmarshal(b, (*E)(event))
	if err != nil {
		return err
	}
	event.Event = newEvent(
		event.I,
		String(event.D),
		func() Time { return event.T },
	)
	return nil
}

type simpleEvent struct {
	i uint64
	t Time
	Content
}

var _ Event = &simpleEvent{}

func (event *simpleEvent) Index() uint64 {
	return event.i
}

func (event *simpleEvent) Time() Time {
	return event.t
}

type jsonMsg struct {
	S   string `json:"session"`
	T   Time   `json:"time"`
	D   string `json:"data"`
	Msg `json:"-"`
}

func newJSONMsg(msg Msg) *jsonMsg {
	if msg == nil {
		return &jsonMsg{}
	}
	return &jsonMsg{
		S: msg.Session(),
		T: msg.Time(),
		D: msg.Text(),
	}
}

func (msg *jsonMsg) MarshalJSON() ([]byte, error) {
	type M jsonMsg
	return json.Marshal((*M)(msg))
}

func (msg *jsonMsg) UnmarshalJSON(b []byte) error {
	type M jsonMsg
	err := json.Unmarshal(b, (*M)(msg))
	if err != nil {
		return err
	}
	msg.Msg = newMsg(msg.S, String(msg.D), func() Time { return msg.T })
	return nil
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
