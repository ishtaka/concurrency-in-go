package main

import "fmt"

func main() {
	generator := func(done <-chan interface{}, integers ...int) <-chan int {
		intChan := make(chan int, len(integers))
		go func() {
			defer close(intChan)
			for _, i := range integers {
				select {
				case <-done:
					return
				case intChan <- i:
				}
			}
		}()

		return intChan
	}

	multiply := func(done <-chan interface{}, intChan <-chan int, multiplier int) <-chan int {
		multipliedChan := make(chan int)
		go func() {
			defer close(multipliedChan)
			for i := range intChan {
				select {
				case <-done:
					return
				case multipliedChan <- i * multiplier:
				}
			}
		}()

		return multipliedChan
	}

	add := func(done <-chan interface{}, intChan <-chan int, additive int) <-chan int {
		addedChan := make(chan int)
		go func() {
			defer close(addedChan)
			for i := range intChan {
				select {
				case <-done:
					return
				case addedChan <- i + additive:
				}
			}
		}()

		return addedChan
	}

	done := make(chan interface{})
	defer close(done)

	intChan := generator(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, intChan, 2), 1), 2)

	for v := range pipeline {
		fmt.Println(v)
	}
}
