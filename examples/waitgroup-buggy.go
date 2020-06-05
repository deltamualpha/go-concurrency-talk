package main

import (
	"fmt"
	"sync"
)

// START OMIT

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(i)
		}()
	}
	// wait for all goroutines to finish.
	fmt.Println("Started all routines")
	wg.Wait()
	fmt.Println("Done")
}

// END OMIT
