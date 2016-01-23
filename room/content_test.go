package room

import "testing"

func TestEvent(t *testing.T) {
	dt := &dumbTime{}
	c1 := String("test content")
	e1 := newEvent(c1, dt.Now)
	if e1.Text() != c1.Text() {
		t.Errorf("content: %v", e1.Text())
	}
	if e1.Time().(t64) != 1 {
		t.Errorf("time: %v", e1.Time())
	}
}

func TestMsg(t *testing.T) {
	dt := &dumbTime{}
	sess := "test session"
	c1 := String("test content")
	m1 := newMsg(sess, c1, dt.Now)
	if m1.Session() != sess {
		t.Errorf("session: %v", m1.Session())
	}
	if m1.Text() != c1.Text() {
		t.Errorf("content: %v", m1.Text())
	}
	if m1.Time().(t64) != 1 {
		t.Errorf("time: %v", m1.Time())
	}
}
