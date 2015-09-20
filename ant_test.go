package ant

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func sleep() { time.Sleep(time.Duration(rand.Intn(1000))) }

type n []int

func (n *n) add(x int) {
	*n = append(*n, x)
}

type z struct{}

func (z z) Error() string { return "z{}" }

func TestFunc(t *testing.T) {
	c := Func(func(s Signal) {
		for i := 0; i < 10; i++ {
			s.Send(rand.Intn(100))
		}
	})

	c = Filter(c, func(t T) bool {
		n, ok := t.(int)
		return ok && n%2 == 1
	})

	//	c = Filter(c, func(n int) bool { return n%2 == 1 })

	for n := range c.Value() {
		t.Log(n)
	}
}
