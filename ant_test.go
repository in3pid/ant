package ant

import (
	"math/rand"
	"testing"
	"time"
)

func sleep() { time.Sleep(1000) }

func TestValues(t *testing.T) {
	c := Func(func(s Signal) {
		for i := 0; i < 10; i++ {
			s.Send(rand.Intn(100))
		}
	})
	c = Filter(c, func(t T) bool {
		n, ok := t.(int)
		return ok && n%2 == 1
	})
	for n := range c.Value() {
		t.Log(n)
	}
}
