package ant

import (
	"math/rand"
	"time"
)

// Func starts a new cursor of f and a new signal, which is
// closed once the funcion returns.
func Func(f func(Signal) error) Signal {
	return Do(signalFunc{NewSignal(), f})
}

type signalFunc struct {
	Signal
	f func(Signal) error
}

func (s signalFunc) Do() { SendErr(s, s.f(s.Signal)) }

// Values starts a cursor sending the argument values.
func Values(values ...Value) Signal {
	return Func(func(s Signal) error {
		for _, v := range values {
			if !Send(s, v) {
				break
			}
		}
		return nil
	})
}

// func Fn(fn func(Signal)) Signal {
// 	return Do(&doFn{NewSignal(), fn})
// }

// type doFn struct {
// 	Signal
// 	fn func(Signal)
// }

// func (do doFn) do() {
// 	do.fn(do.Signal)
// }

// ant.Func(func(s Signal) { s.Value() <- 1 })

type MapFunc func(Value) (Value, error)

type sigMap struct {
	Signal
	from Signal
	fn   MapFunc
}

func (s *sigMap) Do() {
	var err error
	//	var val interface{}
	for {
		// if val != nil || err != nil && !Send(s, val, err) {
		// 	return
		// }
		select {
		case err = <-s.from.Err():

		case val, ok := <-s.from.Value():
			if !ok {
				return
			}
			if val, err = s.fn(val); val == nil {
				continue
			}
			if err != nil {
				SendErr(s, err)
				return
			}
			if !Send(s, val) {
				return
			}
		}
	}
}

func Map(s Signal, fn MapFunc) Signal {
	return Do(&sigMap{NewSignal(), s, fn})
}

type FoldFunc func(Value) error

func Fold(from Signal, fn FoldFunc) error {
	for {
		select {
		case val, ok := <-from.Value():
			if !ok {
				return nil
			}
			if err := fn(val); err != nil {
				return err
			}
		case err := <-from.Err():
			return err
		}
	}
}

func Throttle(s Signal) Signal {
	return Map(s, func(v Value) (Value, error) {
		time.Sleep(time.Duration(rand.Intn(1000)))
		return v, nil
	})
}
