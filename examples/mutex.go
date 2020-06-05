package main

import (
	"fmt"
	"sync"
)

// START OMIT

// SafeCounter is safe to use concurrently.
type SafeCounter struct {
	V   map[string]int
	mux sync.Mutex
}

// Inc increments the counter for the given key.
func (c *SafeCounter) Inc(key string) {
	c.mux.Lock()
	// Lock so only one goroutine at a time can access the map.
	defer c.mux.Unlock()
	c.V[key]++
	fmt.Println(c.V["key"])
}

//END OMIT

func main() {
	c := &SafeCounter{V: make(map[string]int)}
	s := sync.WaitGroup{}
	s.Add(3)
	go func() { c.Inc("key"); s.Done() }()
	go func() { c.Inc("key"); s.Done() }()
	go func() { c.Inc("key"); s.Done() }()
	s.Wait()
}
