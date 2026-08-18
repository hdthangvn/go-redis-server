package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	value int
	mutex sync.Mutex
}

func (c *Counter) Increment() {
	c.mutex.Lock()
	c.value++
	c.mutex.Unlock()
}

func main() {
	counter := Counter{}

	var wg sync.WaitGroup

	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			counter.Increment()
			wg.Done()
		}()
	}

	wg.Wait()

	fmt.Println(counter.value)
}
