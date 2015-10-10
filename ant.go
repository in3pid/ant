package ant

import (
	"fmt"
	"log"
	"reflect"
)

type LogFunc func(fmt string, args ...interface{})

var Logger LogFunc = log.Printf

func antLog(f string, args ...interface{}) { Logger(f, args...) }

type Value interface{}

// A Cursor is a "lazy list" goroutine tied to a communcation agent.
type Cursor interface {
	Signal
}

// A Signaler exposes its Signal by pointer.
type Signaler interface {
	Signal() Signal
}

// A Queryer specifies some query that can be instantiated with some arguments.
type Queryer interface {
	Query(query interface{}) TypeCurser
}

// Curser is a complete selector and can start a cursor.
type Curser interface {
	Cursor(args ...interface{}) Cursor
}

// TypeCurser is a Curser with an optional type association with an inferred decoding strategy.
type TypeCurser interface {
	Curser
	Type(interface{}) Curser
}

// A Listener can attach fan-out cursors onto a cursor.
type Listener interface {
	Listen() Cursor
	StopListen(Cursor)
}

// Close all argument cursors.
func Close(slice ...Cursor) {
	for _, c := range slice {
		c.Close()
	}
}

// CollectErr collects the cursors errors and sends them to a new channel.
func CollectErr(cursors ...Cursor) <-chan error {
	errs := make(chan error)
	errCases := make([]reflect.SelectCase, len(cursors))
	for i, c := range cursors {
		errCases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c.Err())}
	}

	go func() {
		defer close(errs)
		for 0 < len(errCases) {
			n, val, ok := reflect.Select(errCases)
			if !ok {
				copy(errCases[n:], errCases[n+1:])
				errCases = errCases[:len(errCases)-1]
			} else {
				if err, ok := val.Interface().(error); ok {
					errs <- err
				}
			}
		}
	}()

	return errs
}

// StringBytes converts []byte values to strings in a Blob.
// func StringBytes(m Value) Value {
// 	if m, ok := m.(Blob); ok {
// 		for k, v := range m {
// 			if v, ok := v.([]byte); ok {
// 				m[k] = string(v)
// 			}
// 		}
// 	}
// 	return m
// }

type RetryError error

// Retry wraps the Curser such that its Cursor is restarted whenever
// it signals a RetryError (or net.Error with .Temporary()), in which
// case the Signal remains open
// func Retry(c Curser) Curser {
// 	return retrier{c}
// }

// type retrier struct {
// 	curser Curser
// }

// func (r retrier) Cursor(args ...interface{}) Cursor {
// 	return Func(func(s Signal) {
// 		for {
// 			c := r.curser.Cursor(args...)
// 			s.CopyValues(c)
// 			if err := <-c.Err(); err != nil {
// 				switch err := err.(type) {
// 				case net.Error:
// 					if !err.Temporary() {
// 						return
// 					}
// 					antLog(err)
// 				case RetryError:
// 					antLog(err)
// 				default:
// 					return

// 				}
// 			}
// 		}
// 	})
// }

func PanicGuard(s Signal) {
	switch err := recover().(type) {
	case nil:
	case error:
		SendErr(s, err)
	default:
		SendErr(s, fmt.Errorf("ant.Cursor: %v", err))
	}
}
