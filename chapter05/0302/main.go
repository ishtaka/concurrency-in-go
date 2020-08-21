package main

import (
	"fmt"
	"math/rand"
)

func main() {
	doWork := func(done <-chan interface{}) (<-chan interface{}, <-chan int) {
		heartbeat := make(chan interface{}, 1)
		workChan := make(chan int)
		go func() {
			defer close(heartbeat)
			defer close(workChan)

			for i := 0; i < 10; i++ {
				select {
				case heartbeat <- struct{}{}:
				default:
				}

				select {
				case <-done:
					return
				case workChan <- rand.Intn(10):
				}
			}
		}()

		return heartbeat, workChan
	}

	done := make(chan interface{})
	defer close(done)

	heartbeat, results := doWork(done)
	for {
		select {
		case _, ok := <-heartbeat:
			if ok {
				fmt.Println("pulse")
			} else {
				return
			}
		case r, ok := <-results:
			if ok {
				fmt.Printf("results %v\n", r)
			} else {
				return
			}
		}
	}
}
