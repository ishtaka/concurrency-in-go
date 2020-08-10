package main

import "fmt"

func main() {
	chanOwner := func() <-chan int {
		resultCh := make(chan int, 5)
		go func() {
			defer close(resultCh)
			for i := 0; i <= 5; i++ {
				resultCh <- i
			}
		}()

		return resultCh
	}

	resultCh := chanOwner()
	for result := range resultCh {
		fmt.Printf("Received: %d\n", result)
	}
	fmt.Println("Done receiving!")
}
