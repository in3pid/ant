package ant

import (
	"math/rand"
	"testing"
)

func TestValues(t *testing.T) {
	filter := func(x T) T {
		if n, ok := x.(int); ok && n%2 == 0 {
			return nil
		} else {
			return x
		}
	}
	c := Func(func(s Signal) {
		for i := 0; i < 10; i++ {
			s.Send(rand.Intn(100))
		}
	})
	c = Fold(c, filter)
	for n := range c.Value() {
		t.Log(n)
	}
}
