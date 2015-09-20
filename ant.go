package ant

type T interface{}

// A Cursor is a "lazy list" goroutine tied to a communcation agent.
type Cursor interface {
	Err() <-chan error
	Value() <-chan T
	Close()
}

// A SendCursor opens up writes to the channels.
type SendCursor interface {
	SendErr(error) bool
	Send(T) bool
	Close()
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

// CollectErr collects the cursors errors and sends them to a new
// channel. Currently the cursors are read serially and may block unneccesarily.
func CollectErr(cursors ...Cursor) <-chan error {
	// TODO use reflection selects
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

// StringBytes converts []byte values to strings in a Blob.
func StringBytes(m T) T {
	if m, ok := m.(Blob); ok {
		for k, v := range m {
			if v, ok := v.([]byte); ok {
				m[k] = string(v)
			}
		}
	}
	return m
}

func do(c Cursor) Cursor {
	go func() {
		defer c.Close()
		c.(doer).do()
	}()
	return c
}

type doer interface {
	do()
}
