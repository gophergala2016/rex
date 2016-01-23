package room

import "testing"

func TestDumbTime(t *testing.T) {
	dt := new(dumbTime)
	t1 := dt.Now()
	if t1.(t64) != 1 {
		t.Errorf("t1: %v", t1)
	}
	t2 := dt.Now()
	if t2.(t64) != 2 {
		t.Errorf("t2: %v", t2)
	}
}

func TestT64(t *testing.T) {
	t1 := t64(0xDEADBEEF12345678)
	if t1.String() != "deadbeef12345678" {
		t.Errorf("t1: %v", t1)
	}
}
