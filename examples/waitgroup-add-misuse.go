package main

import (
	"fmt"
	"sync"
)

// START OMIT

func main() {
	r := []int{1, 2, 3, 4}
	var group sync.WaitGroup
	for _, p := range r {
		group.Add(1)
		go func(p int) {
			defer group.Done()
			fmt.Println(p)
		}(p)
		fmt.Println("wait")
		group.Wait() // HL
	}
	fmt.Println("done\n")
	for _, p := range r {
		group.Add(1)
		go func(p int) {
			defer group.Done()
			fmt.Println(p)
		}(p)
	}
	fmt.Println("wait")
	group.Wait() // HL
	fmt.Println("done")
}

// END OMIT
