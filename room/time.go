package room

import (
	"encoding/binary"
	"encoding/hex"
	"sync/atomic"
)

var dt = &dumbTime{}

// Time has not been figured out yet.  It is definitely an abstract time.  I
// would like it if all events and messages had unique identifiers.
type Time interface{}

type dumbTime struct {
	t uint64
}

func (t *dumbTime) Time() Time {
	now := atomic.AddUint64(&t.t, 1)
	return t64(now)
}

type t64 uint64

func (t t64) String() string {
	var buf1 [8]byte
	var buf2 [16]byte
	binary.BigEndian.PutUint64(buf1[:], uint64(t))
	hex.Encode(buf2[:], buf1[:])
	return string(buf2[:])
}
