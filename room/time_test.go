package room

import "testing"

func TestDumbTime(t *testing.T) {
	dt := new(dumbTime)
	t1 := dt.Time()
	if t1.(t64) != 1 {
		t.Errorf("t1: %v", t1)
	}
	t2 := dt.Time()
	if t2.(t64) != 2 {
		t.Errorf("t2: %v", t2)
	}
}
