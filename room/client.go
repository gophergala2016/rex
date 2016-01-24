package room

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/bmatsuo/uuid"
	"golang.org/x/net/context"
)

// Client is an interface to a remote REx server.
type Client struct {
	Host    string
	Port    int
	Handler EventHandler
	HTTP    *http.Client
	Now     func() Time
	Session string
}

// NewClient allocates and returns a new client with its Handler set to h.
func NewClient(h EventHandler) *Client {
	return &Client{Handler: h}
}

// NewClientRestore behaves like NewClient but sets the client's Session to
// session as well.
func NewClientRestore(h EventHandler, session string) *Client {
	return &Client{
		Handler: h,
		Session: session,
	}
}

func (c *Client) http() *http.Client {
	if c.HTTP == nil {
		return http.DefaultClient
	}
	return c.HTTP
}

func (c *Client) url(pathquery string) string {
	if strings.HasPrefix(pathquery, "/") {
		pathquery = pathquery[1:]
	}
	return fmt.Sprintf("http://%s:%d/%s", c.Host, c.Port, pathquery)
}

// events performs a long-poll for events on the server.
func (c *Client) events(ctx context.Context, start int) ([]Event, error) {
	log.Printf("POLLING %d", start)
	pathquery := fmt.Sprintf("/rex/v0/events?start=%d", start)
	resp, err := c.http().Get(c.url(pathquery))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(b))
	}
	var events []Event
	dec := json.NewDecoder(resp.Body)
	ejs := newJSONEvent(nil)
	for {
		*ejs = jsonEvent{}
		err := dec.Decode(&ejs)
		if err != nil {
			break
		}
		events = append(events, ejs.Event)
	}
	if err == io.EOF {
		return events, nil
	}
	return events, err
}

// send sends a message to the remote server with the given session identifier
// (not c.Session).
func (c *Client) send(ctx context.Context, session string, content Content) error {
	_dt := c.Now
	if _dt == nil {
		_dt = dt.Now
	}
	_m := newMsg(session, content, _dt)
	m := newJSONMsg(_m)
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	body := bytes.NewReader(b)
	u := c.url("/rex/v0/messages")
	resp, err := c.http().Post(u, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%v %s %s: %s", resp.Status, "POST", u, b)
	}
	return nil
}

// CreateSession initializes c.Session by registering an identifier with the
// remote bus. If the value is already set no registration is performed.
func (c *Client) CreateSession(ctx context.Context, name string) error {
	if c.Session != "" {
		return nil
	}

	session := uuid.New()
	// FIXME: first message sent is the name of the session? that seems...
	// ~reasonable.
	err := c.send(ctx, session, String(name))
	if err == nil {
		c.Session = session
	}
	return err
}

// Send sends a message to the remote server using the given session
// identifier.
func (c *Client) Send(ctx context.Context, content Content) error {
	if c.Session == "" {
		return fmt.Errorf("no session id")
	}
	return c.send(ctx, c.Session, content)
}

// Run processes events received from the remote bus.
func (c *Client) Run(ctx context.Context, start int) (next int, err error) {
	term := ctx.Done()
evloop:
	for {
		if start < next {
			start = next
		}
		evs, err := c.events(ctx, start)
		if err != nil {
			return next, err
		}
		log.Printf("[INFO] Found %d new events", len(evs))
		select {
		case <-term:
			break evloop
		default:
		}
		for _, ev := range evs {
			if c.Handler != nil {
				c.Handler.HandleEvent(ctx, c, ev)
			}
			select {
			case <-term:
				break evloop
			default:
			}
			next = int(ev.Index()) + 1
		}
	}
	return next, nil
}

// EventHandler is part of the client's event loop.
type EventHandler interface {
	HandleEvent(context.Context, *Client, Event)
}

// ehfunc implements EventHandler
type ehfunc func(context.Context, *Client, Event)

func (fn ehfunc) HandleEvent(ctx context.Context, c *Client, event Event) {
	fn(ctx, c, event)
}
