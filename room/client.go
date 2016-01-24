package room

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bmatsuo/uuid"
	"golang.org/x/net/context"
)

// StatusClientConnected is a type of message that indicates the presence of a
// new client.
const StatusClientConnected = "ClientConnected"

// Client is an interface to a remote REx server.
type Client struct {
	Host    string
	Port    int
	Handler EventHandler
	HTTP    *http.Client
	Session string
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
	resp, err := c.http().Get(c.url("/rex/v0/events"))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(b))
	}
	//var events []Event
	//json.NewDecoder(resp.Body)
	return nil, nil
}

// send sends a message to the remote server with the given session identifier
// (not c.Session).
func (c *Client) send(ctx context.Context, session string, content Content) error {
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
func (c *Client) Run(ctx context.Context, start int, handler EventHandler) error {
	return nil
}

// ClientConfig controls behaviors of a Client object.
type ClientConfig struct {
}

// Dial creates a client session and connects to the specified
func Dial(addr string) *Client {
	return nil
}

// EventHandler is part of the client's event loop.
type EventHandler interface {
	HandleEvent(context.Context, Event)
}

// ehfunc implements EventHandler
type ehfunc func(context.Context, Event)

func (fn ehfunc) HandleEvent(ctx context.Context, event Event) {
	fn(ctx, event)
}
