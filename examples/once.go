package main

import (
	"fmt"
	"sync"
)

// START OMIT

func main() {
	var once sync.Once
	onceBody := func() {
		fmt.Println("Only once")
	}
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			// prints "Only once" once, though Do() is called ten times.
			once.Do(onceBody)
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done // wait for all goroutines to finish.
	}
	fmt.Println("complete")
}

// END OMIT
