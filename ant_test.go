package ant

import (
	"log"
	"math/rand"
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
	c := Func(func(s Signal) error {
		for i := 0; i < 10; i++ {
			if !Send(s, rand.Intn(100)) {
				return nil
			}
		}
		return nil
	})

	c = Map(c, func(t Value) (Value, error) {
		if n, ok := t.(int); ok && n%2 == 1 {
			return n, nil
		}
		return nil, nil
	})

	//	c = Filter(c, func(n int) bool { return n%2 == 1 })

	for n := range c.Value() {
		log.Println(n)
	}
}
