package main

import (
	"sync"
	"testing"
)

func TestCounterIncrement(t *testing.T) {
	counter := Counter{}
	var wg sync.WaitGroup

	const goroutines = 100000
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()

	got := counter.Value()
	if got != goroutines {
		t.Errorf("expected %d, got %d", goroutines, got)
	}
}
