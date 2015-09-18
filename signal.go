package ant

import (
	"time"
)

// Signal is the main communication object and provides the Cursor interface.
type Signal struct {
	stop chan struct{}
	err  chan error
	val  chan T
}

// MakeSignal instantiates a signal.
func MakeSignal() Signal {
	return Signal{stop: make(chan struct{}), err: make(chan error, 1), val: make(chan T)}
}

// SendErr tries to send the error, returning success. False means the
// cursor is closed. It wont send nil and returns true on nil arguments.
func (s Signal) SendErr(err error) bool {
	if err == nil {
		return false
	}
	select {
	case <-s.stop:
		return false
	case s.err <- err:
		return true
	}
}

// Send tries to snd the value, returning success. False means the
// cursor is closed. It wont sent nil and returns true
// on nil arguments.
func (s Signal) Send(value T) bool {
	select {
	case <-s.stop:
		return false
	case s.val <- val:
		return true
	}
}

// Err returns the error channel for reading.
func (s Signal) Err() <-chan error {
	return s.err
}

// Value returns the value channel for reading.
func (s Signal) Value() <-chan T {
	return s.val
}

// Close notifies the goroutine on the stop channel and closes the signal.
func (s Signal) Close() {
	close(s.stop)
	time.Sleep(100)
	close(s.err)
	close(s.val)
}

// CopyErr copies the error, if any, from c and returns true if any
// error was ent.
func (s Signal) CopyErr(c Cursor) bool {
	err, ok := <-c.Err()
	if ok {
		s.SendErr(err)
	}
	return ok
}
