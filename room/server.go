package room

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Room represents a single shared enivornment managed by a server.  The
// service is advertised using mDNS an must conform to the format specified in
// RFC 6763 Section 7.  The Name may contain any unicode text excluding ASCII
// control characters but is recommended to not contain '\n' bytes for display
// purposes.  An mDNS instance identifier will be generated from the given
// name, the time and the process identifier.
type Room struct {
	Name    string
	Service string
}

// ServerConfig controls how a server advertises itself to potential clients as
// well as miscelaneous communication behaviors.
type ServerConfig struct {
	Room *Room
	Bus  *Bus

	// Addr is an optional address to bind.  If empty, the address of ":0" will
	// be used.
	Addr string
}

// Server is a server used by a TV application to run a game or collaborative
// procedure.
type Server struct {
	config   *ServerConfig
	handler  *httpBus
	tcp      *net.TCPListener
	http     *http.Server
	serving  chan struct{}
	serveErr chan error
}

// NewServer initializes a new server, but does not start serving clients.
func NewServer(config *ServerConfig) *Server {
	if config == nil {
		panic("nil config")
	}

	s := &Server{}
	s.config = config
	s.initHTTP()

	return s
}

func (s *Server) initHTTP() {
	if s.handler != nil {
		panic("already initialized")
	}
	s.handler = newHTTPBus(s.bus())
	s.serving = make(chan struct{})
	s.serveErr = make(chan error, 1)
	s.http = &http.Server{
		Addr:         s.config.Addr, // FIXME not correct
		Handler:      s.handler,
		ReadTimeout:  250 * time.Millisecond,
		WriteTimeout: 0,
	}
}

func (s *Server) bus() *Bus {
	return s.config.Bus
}

// Start binds the server to a port and beigns allowing clients to connect.
// Start must not be called more than once.
func (s *Server) Start() error {
	go func() {
		defer close(s.serveErr)
		err := s.listenTCP()
		if err != nil {
			s.serveErr <- err
			return
		}
		s.serveErr <- nil
		s.serveErr <- s.http.Serve(s.tcp)
	}()

	err := <-s.serveErr
	if err == nil {
		// signal that we are serving clients
		close(s.serving)
	}
	return err
}

// Wait returns when the server has terminated.  Wait must not be called more
// than once.
func (s *Server) Wait() error {
	select {
	case <-s.serving:
		// we have started serving traffic... righteous.
	default:
		panic("Wait called before Start")
	}
	err, ok := <-s.serveErr
	if !ok {
		return fmt.Errorf("too main goroutines waiting")
	}
	return err
}

// Run binds to a random port, begins broadcasting service metadata using mDNS,
// and begins streaming client events and dispatching client messages.
// Typically, Run never returns a value. If any critical error is encountered
// it will be returned.
//
// Run is equivalent to calling Start, followed by s.Wait.
func (s *Server) Run() error {
	err := s.Start()
	if err != nil {
		return err
	}
	return s.Wait()
}

func (s *Server) listenTCP() error {
	var err error
	addr := s.config.Addr
	laddr := &net.TCPAddr{}
	if addr != "" {
		host, _port, err := net.SplitHostPort(addr)
		if err != nil {
			return err
		}
		laddr.Port, err = strconv.Atoi(_port)
		if err != nil {
			return fmt.Errorf("invalid address port")
		}
		laddr.IP = net.ParseIP(host)
		if laddr.IP == nil && host != "" {
			return fmt.Errorf("invalid address host")
		}
	}

	s.tcp, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	return nil
}

// Addr returns the string address the bus is listening on for HTTP requests.
func (s *Server) Addr() string {
	return s.tcp.Addr().String()
}

// Event broadcasts c to all connected clients, giving it the next unused event
// index.
func (s *Server) Event(c Content) {
	s.bus().Event(c)
}

func newBusHandler(b *Bus) http.Handler {
	return newHTTPBus(b)
}

// httpBus exposes the bus functions Subscribe and Message over http endpoints.
type httpBus struct {
	b   *Bus
	mux *http.ServeMux // FIXME use something that is faster
}

func newHTTPBus(b *Bus) *httpBus {
	h := &httpBus{
		b:   b,
		mux: http.NewServeMux(),
	}

	// register all api routes
	h.mux.HandleFunc("/rex/v0/events", busEventsHandler(b))
	h.mux.HandleFunc("/rex/v0/messages", busMessagesHandler(b))
	// TODO: a way for new clients to catch up without log compaction
	// h.mux.HandleFunc("/rex/v0/state", busStateHandler(b))

	return h
}

func jsonError(id, reason string) string {
	return fmt.Sprintf(`{"error":%q, "reason":%q}`, id, reason)
}

func jsonMethodNotAllowed(allow ...string) string {
	return jsonError("http_method_invalid", fmt.Sprintf("request method must be one of %v", allow))
}

func busEventsHandler(b *Bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, jsonMethodNotAllowed)
			return
		}

		q := r.URL.Query()
		_start := q.Get("start")
		start := 0
		if _start != "" {
			var err error
			start, err = strconv.Atoi(_start)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, jsonError("parameter_invalid", "invalid start index"))
				return
			}
		}

		sub := b.Subscribe(start)
		defer b.Unsubscribe(sub)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		var timeout <-chan time.Time
		for sub.Next(timeout) {
			if timeout == nil {
				timeout = time.After(time.Millisecond)
			}
			event := sub.Event()
			ejs := newJSONEvent(event)
			err := enc.Encode(ejs)
			if err != nil {
				log.Printf("[INFO] failed to deliver event to client: %v", err)
				return
			}
		}
	}
}

func busMessagesHandler(b *Bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, jsonMethodNotAllowed)
			return
		}

		msg := map[string]interface{}{}
		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			var resp string
			switch e := err.(type) {
			case *json.SyntaxError:
				resp = e.Error()
			default:
				log.Printf("[INFO] message i/o error: %v", err)
				resp = "could not read a complete entity"
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, jsonError("http_request_invalid", resp))
			return
		}

		log.Printf("[INFO] message received %v", msg)

		var content string
		_content, ok := msg["data"]
		if ok {
			content, ok = _content.(string)
		}
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, jsonError("protocol_error", "missing message content"))
			return
		}
		var session string
		_session, ok := msg["session"]
		if ok {
			session, ok = _session.(string)
		}
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, jsonError("protocol_error", "missing message content"))
			return
		}

		b.Message(session, String(content))
	}
}

func (b *httpBus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.mux.ServeHTTP(w, r)
}
