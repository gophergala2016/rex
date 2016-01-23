package room

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHTTPBusEvents(t *testing.T) {
	b := newBus(nil)
	defer b.close()

	go func() {
		time.Sleep(time.Second)
		for i := 0; i < 100; i++ {
			b.Event(String("test content"))
		}
	}()

	h := newBusHandler(b)
	s := httptest.NewServer(h)
	defer s.Close()

	url := fmt.Sprintf("%s/rex/v0/events", s.URL)
	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("http: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("error %s: %s", resp.Status, b)
		return
	}

	dec := json.NewDecoder(resp.Body)
	for {
		e := map[string]interface{}{}
		err := dec.Decode(&e)
		if err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		content, ok := e["data"].(string)
		if !ok || content != "test content" {
			t.Errorf("event: %v", e)
		}
		return
	}
}

func TestHTTPBusMessages(t *testing.T) {
	msglock := &sync.Mutex{}
	msg := []Msg{}
	hmsg := func(m Msg) {
		msglock.Lock()
		msg = append(msg, m)
		msglock.Unlock()
	}

	b := newBus(hmsg)
	defer b.close()

	h := newBusHandler(b)
	s := httptest.NewServer(h)
	defer s.Close()

	url := fmt.Sprintf("%s/rex/v0/messages", s.URL)
	resp, err := http.Post(url, "application/json", strings.NewReader(`{
		"session": "session-01",
		"time": "0000010000000001",
		"data": "test content"
	}`))
	if err != nil {
		t.Errorf("http: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("error %s: %s", resp.Status, b)
		return
	}

	time.Sleep(50 * time.Millisecond)

	msglock.Lock()
	n := len(msg)
	msglock.Unlock()

	if n != 1 {
		t.Errorf("messages: %d", n)
		return
	}

	if msg[0].Text() != "test content" {
		t.Errorf("content: %v", msg[0].Text())
	}
}
