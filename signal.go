package ant

import (
	"time"
)

type Signal struct {
	stop chan struct{}
	err  chan error
	val  chan T
}

func MakeSignal() Signal {
	return Signal{
		stop: make(chan struct{}, 1),
		err:  make(chan error, 1),
		val:  make(chan T)}
}

func (s Signal) SendErr(err error) bool {
	select {
	case <-s.stop:
		return false
	case s.err <- err:
		return true
	}
}

func (s Signal) Send(val T) bool {
	select {
	case <-s.stop:
		return false
	case s.val <- val:
		return true
	}
}

func (s Signal) Err() <-chan error {
	return s.err
}

func (s Signal) Value() <-chan T {
	return s.val
}

func (s Signal) Close() {
	close(s.stop)
	time.Sleep(100)
	close(s.err)
	close(s.val)
}

func (s Signal) CopyErr(c Cursor) bool {
	err, ok := <-c.Err()
	if ok {
		s.SendErr(err)
	}
	return ok
}
