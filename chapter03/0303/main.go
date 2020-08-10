package main

import "fmt"

func main() {
	intCh := make(chan int)
	go func() {
		defer close(intCh)

		for i := 1; i <= 5; i++ {
			intCh <- i
		}
	}()

	for integer := range intCh {
		fmt.Printf("%v ", integer)
	}
}
