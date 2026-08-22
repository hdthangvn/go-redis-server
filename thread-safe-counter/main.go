package main

import (
	"fmt"
	"sync"
)

type Counter struct {
	mutex sync.Mutex
	value int
}

func (c *Counter) Value() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.value
}

// Nếu dòng nằm giữa Lock() và Unlock() bị lỗi hoặc panic, mutex ko gọi đc Unlock() -> bị deadLock.
// Giải pháp dùng defer để đảm bảo Unlock() luôn được gọi.
func (c *Counter) Increment() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value++
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

	fmt.Println(counter.Value())
}
