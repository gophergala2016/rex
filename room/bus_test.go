package room

import (
	"testing"
	"time"
)

func TestBus(t *testing.T) {
	b := newBus(nil)
	b.close()
}

func TestSubscription(t *testing.T) {
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
