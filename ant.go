package ant

type T interface{}

// Cursor is a value iterator goroutine tied to a communication agent.
type Cursor interface {
	Value() <-chan T
	Err() <-chan error
	Close()

	do()
}

// Queryer specifies some selection query with optional type associations.
type Queryer interface {
	Query(query interface{}) TypeCurser
}

// TypeCurser is a Curser with a type selector..
type TypeCurser interface {
	Curser

	Map() Curser
	Struct(interface{}) Curser
}

// Curser is a complete selector and can start a Cursor.
type Curser interface {
	Cursor(args ...interface{}) Cursor
}

//FanIn multiplexes a set of cursors onto a single one.
// func FanIn(cursors ...cursor) Cursor {
// 	return nil
// }

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

// StringBytes retype []byte
// values to strings in a map[string]interface{}
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
