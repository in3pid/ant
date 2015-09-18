package ant

type T interface{}

type Cursor interface {
	do()

	Close()
	Err() <-chan error
	Value() <-chan T
}

type Curser interface {
	Cursor(...interface{}) Cursor
}

type Queryer interface {
	Query(query interface{}) TypeCurser
}

type TypeCurser interface {
	Curser

	Map() Curser
	Struct(interface{}) Curser
}

// func Exec(q Queryer, c Cursor) error {
// 	for r := range c.Row() {
// 		if _, err := q.Exec(r); err != nil {
// 			c.SendStop()
// 			return err
// 		}
// 	}
// 	return <-c.Err()
// }

// CollectErr collects all cursor errors to a single channel. It
// collects serially and may block unneccesarily.
func CollectErr(cursors ...Cursor) <-chan error {
	errs := make(chan error)
	go func() {
		defer close(errs)
		for _, c := range cursors {
			if err, ok := <-c.Err(); ok {
				errs <- err
			}
		}
	}()
	return errs
}

// StringBytes mutates a map[string]interface{} converting each []byte value to string.
func StringBytes(m T) T {
	if m, ok := m.(map[string]T); ok {
		for k, v := range m {
			if v, ok := v.([]byte); ok {
				m[k] = string(v)
			}
		}
	}
	return m
}

// Values returns a cursor populated with the values.
func Values(values ...T) Cursor {
	return Func(func(s Signal) {
		for _, v := range values {
			s.Send(v)
		}
	})
}

// Func instantiates a new cursor consisting of a fresh Signal and the
// func in a goroutine. The signal is assuredly closed.
func Func(f func(Signal)) Cursor {
	return do(signalFunc{MakeSignal(), f})
}
func (c signalFunc) do() { c.f(c.Signal) }

type signalFunc struct {
	Signal
	f func(Signal)
}

// Fold each value of the cursor through a list of functions, possibly
// mutating, passing the result to upstream. If any return nil this
// value is skipped.
func Fold(c Cursor, f ...func(T) T) Cursor {
	if fc, ok := c.(*foldCursor); ok {
		fc.f = append(fc.f, f...)
		return c
	}
	return do(&foldCursor{MakeSignal(), c, f})
}

func do(c Cursor) Cursor {
	go func() {
		defer c.Close()
		c.do()
	}()
	return c
}

type foldCursor struct {
	Signal
	c Cursor
	f []func(T) T
}

func (c foldCursor) do() {
loop:
	for v := range c.c.Value() {
		for _, f := range c.f {
			if v = f(v); v == nil {
				continue loop
			}
		}
		if !c.Send(v) {
			break
		}
	}
	c.CopyErr(c.c)
}
