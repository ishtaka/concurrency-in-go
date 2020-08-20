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

	bridge := func(done <-chan interface{}, chanChan <-chan <-chan interface{}) <-chan interface{} {
		valChan := make(chan interface{})
		go func() {
			defer close(valChan)
			for {
				var stream <-chan interface{}
				select {
				case maybeChan, ok := <-chanChan:
					if ok == false {
						return
					}
					stream = maybeChan
				case <-done:
					return
				}
				for val := range orDone(done, stream) {
					select {
					case valChan <- val:
					case <-done:
					}
				}
			}
		}()

		return valChan
	}

	genVals := func() <-chan <-chan interface{} {
		chanChan := make(chan (<-chan interface{}))
		go func() {
			defer close(chanChan)
			for i := 0; i < 10; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				chanChan <- stream
			}
		}()

		return chanChan
	}

	for v := range bridge(nil, genVals()) {
		fmt.Printf("%v ", v)
	}
}
