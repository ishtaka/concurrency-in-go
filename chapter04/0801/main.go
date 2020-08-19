package main

import "fmt"

func main() {
	orDone := func(done, c <-chan interface{}) <-chan interface{} {
		valChan := make(chan interface{})
		go func() {
			defer close(valChan)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if ok == false {
						return
					}
					select {
					case valChan <- v:
					case <-done:
					}
				}
			}
		}()

		return valChan
	}

	done := make(chan interface{})
	defer close(done)

	test := make(chan interface{})
	go func() {
		defer close(test)
		for i := 1; i <= 3; i++ {
			select {
			case <-done:
				return
			case test <- i:
			}
		}
	}()

	for val := range orDone(done, test) {
		fmt.Println(val)
	}
}
