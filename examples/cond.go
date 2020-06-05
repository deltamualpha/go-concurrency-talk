package main

import (
	"fmt"
	"sync"
)

// START OMIT

func main() {
	var v int
	m := sync.Mutex{}
	m.Lock() // 1. main process is owner of lock.
	c := sync.NewCond(&m)
	go func() {
		m.Lock() // 4. blocks until Wait(); goroutine then owns lock.
		v = 1
		c.Broadcast() // 5. let waiters know we're done.
		m.Unlock()    // 6. goroutine has released lock.
	}()
	v = 0          // 2. do some initialization.
	c.Wait()       // 3. Temporarily release the lock and block until Broadcast() is called.
	fmt.Println(v) // 7. prints "1".
	m.Unlock()
}

// END OMIT
