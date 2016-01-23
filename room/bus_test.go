package room

import (
	"testing"
	"time"
)

func TestBus(t *testing.T) {
	b := newBus(nil)
	b.close()
}

func TestBusMessage(t *testing.T) {
	done := make(chan struct{})
	defer func() {
		timeout := time.After(time.Second)
		select {
		case <-timeout:
			t.Errorf("timeout terminating")
		case <-done:
		}
	}()

	m := make(chan Msg)
	h := func(msg Msg) {
		defer close(done)
		timeout := time.After(time.Second)
		select {
		case <-timeout:
			t.Errorf("timeout delivering message")
		case m <- msg:
		}
	}

	session := "test session"
	content := String("test message")
	b := newBus(h)
	defer b.close()
	b.Message(session, content)

	timeout := time.After(time.Second)
	select {
	case <-timeout:
		t.Errorf("timeout delivering message")
	case msg := <-m:
		if msg.Session() != session {
			t.Errorf("session: %v", msg.Session())
		}
		if msg.Text() != content.Text() {
			t.Errorf("content: %v", msg.Text())
		}
	}
}

func TestBusSubscription(t *testing.T) {
	b := newBus(nil)

	content := String("this is a test")

	go func() {
		b.Event(content)
		time.Sleep(10 * time.Millisecond)
		b.close()
	}()

	s := b.Subscribe(0)
	defer b.Unsubscribe(s)

	n := 0
	for s.Next() {
		n++
		event := s.Event()
		if event.Text() != content.Text() {
			t.Errorf("content: %v", event.Text())
		}
	}
	if n != 1 {
		t.Errorf("num event: %d", n)
	}
}
