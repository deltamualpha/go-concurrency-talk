package main

import (
	"fmt"
	"sync"
)

// START OMIT

// SafeCounter is safe to use concurrently.
type SafeCounter struct {
	v   map[string]int
	mux sync.RWMutex
}

// Inc increments the counter for the given key.
func (c *SafeCounter) Inc(key string) {
	c.mux.Lock()
	// Lock so only one goroutine at a time can access the map.
	defer c.mux.Unlock()
	c.v[key]++
}

// Value returns the current value of the counter for the given key.
func (c *SafeCounter) Value(key string) int {
	c.mux.RLock()
	// Many readers can get this lock at once, but Lock() takes priority.
	defer c.mux.RUnlock()
	return c.v[key]
}

//END OMIT

func main() {
	c := &SafeCounter{v: make(map[string]int)}
	s := sync.WaitGroup{}
	s.Add(3)
	go func() {
		c.Inc("key")
		fmt.Println(c.Value("key"))
		s.Done()
	}()
	go func() {
		c.Inc("key")
		fmt.Println(c.Value("key"))
		s.Done()
	}()
	go func() {
		c.Inc("key")
		fmt.Println(c.Value("key"))
		s.Done()
	}()
	s.Wait()
}
