package ant

// Func starts a new cursor from f and a new Signal. The signal is
// assuredly closed afterwards.
func Func(f func(Signal)) Cursor {
	return do(signalFunc{MakeSignal(), f})
}

type signalFunc struct {
	Signal
	f func(Signal)
}

func (c signalFunc) do() { c.f(c.Signal) }

// Values starts a cursor sending the argument values.
func Values(values ...T) Cursor {
	return Func(func(s Signal) {
		for _, v := range values {
			if !s.Send(v) {
				break
			}
		}
	})
}

// Fold the cursor such that for each value v, Fn(Fn-1(...(v))) is
// passed to upstream. The value is skipped if any return nil
func Fold(c Cursor, f ...func(T) T) Cursor {
	if fc, ok := c.(*foldCursor); ok {
		fc.f = append(fc.f, f...)
		return c
	}
	return do(&foldCursor{MakeSignal(), c, f})
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
