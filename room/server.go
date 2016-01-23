package room

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func newBusHandler(b *bus) http.Handler {
	return newHTTPBus(b)
}

// httpBus exposes the bus functions Subscribe and Message over http endpoints.
type httpBus struct {
	b   *bus
	mux *http.ServeMux // FIXME use something that is faster
}

func newHTTPBus(b *bus) *httpBus {
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

func busEventsHandler(b *bus) http.HandlerFunc {
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

		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		for sub.Next() {
			event := sub.Event()
			m := map[string]interface{}{
				"index": event.Index(),
				"time":  event.Time(),
				"data":  event.Text(),
			}
			err := enc.Encode(m)
			if err != nil {
				log.Printf("[INFO] failed to deliver event to client: %v", err)
				return
			}
		}
	}
}

func busMessagesHandler(b *bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, jsonMethodNotAllowed)
			return
		}

		msg := map[string]interface{}{}
		err := json.NewDecoder(r.Body).Decode(msg)
		if err != nil {
			var resp string
			switch e := err.(type) {
			case *json.SyntaxError:
				resp = e.Error()
			default:
				resp = "could not read a complete entity"
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, jsonError("http_request_invalid", resp))
			return
		}

		log.Printf("[INFO] message received %v", msg)
	}
}

func (b *httpBus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.mux.ServeHTTP(w, r)
}
