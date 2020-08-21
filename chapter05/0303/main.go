package main

import (
	"time"
)

func DoWork(done <-chan interface{}, nums ...int) (<-chan interface{}, <-chan int) {
	heartbeat := make(chan interface{}, 1)
	intChan := make(chan int)
	go func() {
		defer close(heartbeat)
		defer close(intChan)

		time.Sleep(2 * time.Second)

		for _, n := range nums {
			select {
			case heartbeat <- struct{}{}:
			default:
			}

			select {
			case <-done:
				return
			case intChan <- n:
			}
		}
	}()

	return heartbeat, intChan
}

func main() {
}
