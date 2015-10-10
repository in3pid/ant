package ant

import (
	"fmt"
	"runtime"
)

// Signal is the main communication object and provides the Cursor interface.
// type Signal struct {
// 	stop chan struct{}
// 	err  chan error
// 	val  chan Value
// }

// // MakeSignal instantiates a signal.
// func MakeSignal() Signal {
// 	return Signal{stop: make(chan struct{}), err: make(chan error, 1), val: make(chan Value)}
// }

// // ewSignal instantiates a signal.
// func NewSignal() *Signal {
// 	return &MakeSignal()
// }

// SendErr tries to send the error, returning success. False means the
// cursor is closed. It wont send nil and returns true on nil arguments.
// func (s Signal) SendErr(err error) bool {
// 	if err == nil {
// 		return false
// 	}
// 	select {
// 	case <-s.stop:
// 		return false
// 	case s.err <- err:
// 		return true
// 	}
// }

// // Send tries to snd the value, returning success. False means the
// // cursor is closed. It wont sent nil and returns true
// // on nil arguments.
// func (s Signal) Send(value Value) bool {
// 	if value == nil {
// 		return true
// 	}
// 	select {
// 	case <-s.stop:
// 		return false
// 	case s.val <- value:
// 		return true
// 	}
// }

// // Stop returns the stop signal channel for reading.
// func (s Signal) Stop() <-chan struct{} {
// 	return s.stop
// }

// // Err returns the error channel for reading.
// func (s Signal) Err() <-chan error {
// 	return s.err
// }

// // Value returns the value channel for reading.
// func (s Signal) Value() <-chan Value {
// 	return s.val
// }

// // Close notifies the goroutine on the stop channel and closes the signal.
// func (s Signal) Close() {
// 	close(s.stop)
// 	runtime.Gosched()
// 	close(s.err)
// 	close(s.val)
// }

// // Closed answers whether this signal is closed or open.
// func (s Signal) Closed() bool {
// 	select {
// 	case <-s.stop:
// 		return true
// 	default:
// 		return false
// 	}
// }

// func (s *Signal) Signal() *Signal { return s }

// CopyErr copies the errors until either Signal closes and returns
// the .
func CopyErr(to, from Signal) (n int) {
	for err := range from.Err() {
		if n++; !SendErr(to, err) {
			break
		}
	}
	return n
}

// CopyValues copies the values until either Signal closes and returns
// the count.
func CopyValues(to, from Signal) (n int) {
	for v := range from.Value() {
		if n++; !Send(to, v) {
			break
		}
	}
	return n
}

// NewLogger creates a logging Signal. It takes 0 to 2 arguments. 2)
// values are logged with p[0] and errors with p[1]. 1) errors
// are logged with p[0]. 0) errors are logged with log.Printf.
func NewLogger(val, err LogFunc) Signal {
	return Do(&logger{NewSignal(), val, err})
}

func guard(s Signal) {
	if e := recover(); e != nil {
		err, ok := e.(error)
		if !ok {
			err = fmt.Errorf("ant: %v", e)
		}
		SendErr(s, err)
	}
	s.Close()
}

func Do(s Doer) Signal {
	go func() {
		defer guard(s)
		s.Do()
	}()
	return s
}

type Doer interface {
	Signal
	Do()
}

type logger struct {
	Signal
	logv LogFunc
	loge LogFunc
}

func (l *logger) Do() {
	val, err := l.Value(), l.Err()
	for val != nil || err != nil {
		select {
		case v, ok := <-val:
			if !ok {
				val = nil
			} else {
				if l.logv != nil {
					l.logv("%v", v)
				}
			}

		case e, ok := <-err:
			if !ok {
				err = nil
			} else {
				if l.loge != nil {
					l.loge("%v", e)
				}
			}
		}
	}
}

func Copy(to, from Signal) (nValue, nErr int) {
	val, err := from.Value(), from.Err()
	for !(val == nil && err == nil) {
		select {
		case e, ok := <-err:
			if nErr++; !(ok && SendErr(to, e)) {
				err = nil
			}
		case v, ok := <-val:
			if nValue++; !(ok && Send(to, v)) {
				val = nil
			}
		}
	}
	return nValue, nErr
}

type callbackSig struct {
	Signal
	onValue func(Value)
	onErr   func(error)
}

func (s callbackSig) SendErr(err error) bool {
	if s.onErr != nil {
		s.onErr(err)
	}
	return SendErr(s.Signal, err)
}

func (s callbackSig) SendValue(v Value) bool {
	if s.onValue != nil {
		s.onValue(v)
	}
	return Send(s.Signal, v)
}

func Callback(s Signal, onValue func(Value), onErr func(error)) Signal {
	return &callbackSig{s, onValue, onErr}
}

type Signal interface {
	Stop() chan struct{}
	Value() chan Value
	Err() chan error
	Close()
}

func NewSignal() Signal {
	return &signal{stop: make(chan struct{}), err: make(chan error, 1), val: make(chan Value)}
}

type signal struct {
	stop chan struct{}
	err  chan error
	val  chan Value
}

// SendErr tries to send the error, returning success. False means the
// cursor is closed. It wont send nil and returns true on nil arguments.
func SendErr(s Signal, err error) bool {
	if err == nil {
		return true
	}
	select {
	case <-s.Stop():
		return false
	case s.Err() <- err:
		return true
	}
}

// Send tries to snd the value, returning success. False means the
// cursor is closed. It wont sent nil and returns true
// on nil arguments.
func Send(s Signal, value Value) bool {
	if value == nil {
		return true
	}
	select {
	case <-s.Stop():
		return false
	case s.Value() <- value:
		return true
	}
}

// Stop returns the stop channel.
func (s *signal) Stop() chan struct{} {
	return s.stop
}

// Err returns the error  channel.
func (s *signal) Err() chan error {
	return s.err
}

// Value returns the value channel.
func (s *signal) Value() chan Value {
	return s.val
}

// Close notifies the goroutine on the stop channel and closes the signal.
func (s *signal) Close() {
	close(s.stop)
	runtime.Gosched()
	close(s.err)
	close(s.val)
}

// Closed answers whether this signal is closed or open.
func Closed(s Signal) bool {
	select {
	case <-s.Stop():
		return true
	default:
		return false
	}
}
