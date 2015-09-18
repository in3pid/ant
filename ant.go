package ant

type T interface{}

// Cursor is a value iterator goroutine tied to a communication agent.
type Cursor interface {
	Value() <-chan T
	Err() <-chan error
	Close()

	do()
}

// Queryer instantiates a Curser with type selection.
type Queryer interface {
	Query(query interface{}) TypeCurser
}

// TypeCurser may select a different decoder.
type TypeCurser interface {
	Curser

	Map() Curser
	Struct(interface{}) Curser
}

// Curser is a fully setup selector and can start a cursor.
type Curser interface {
	Cursor(args ...interface{}) Cursor
}

//FanIn multiplexes a set of cursors onto a single one.
func FanIn(cursors ...cursor) Cursor {

}

// Filter is a predicate filter for Fold.
func Filter(f func(T) bool) func(T) T {
	return func(t T) T {
		if f(t) {
			return t
		} else {
			return nil
		}
	}
}

// CollectErr fans through the cursors, serially, and sends their errors to a new channel. May block unneccesarily but this is subject
// to change.
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

// StringBytes mutates a map[string]interface{} retyping []byte
// values to strings
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

func do(c Cursor) Cursor {
	go func() {
		defer c.Close()
		c.do()
	}()
	return c
}
