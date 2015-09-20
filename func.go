package ant

// Func starts a new cursor of f and a new signal, which is
// closed once the funcion returns.
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

type FoldFunc func(T) T

// Fold the cursor such that for each value v, Fn(Fn-1(...(v))) is
// passed to upstream. The value is skipped if any function return nil.
func Fold(c Cursor, f ...func(T) T) Cursor {
	if fc, ok := c.(*foldCursor); ok {
		fc.f = append(fc.f, f...)
		return c
	} else {
		return do(&foldCursor{MakeSignal(), c, f})
	}
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

// Filter c by the predicate f.
func Filter(c Cursor, f func(T) bool) Cursor {
	return Fold(c, filter(f))
}

func filter(f func(T) bool) func(T) T {
	return func(t T) T {
		if f(t) {
			return t
		} else {
			return nil
		}
	}
}

// ForwardFilter emits the filtered values to a channel.
func ForwardFilter(to chan<- T, c Cursor, f func(T) bool) Cursor {
	return Fold(c, forwardFilter(to, f))
}

func forwardFilter(to chan<- T, f func(T) bool) func(T) T {
	return func(t T) T {
		if f(t) {
			return t
		} else {
			to <- t
			return nil
		}
	}
}
