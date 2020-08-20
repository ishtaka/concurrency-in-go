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

	tee := func(done <-chan interface{}, in <-chan interface{}) (_, _ <-chan interface{}) {
		out1 := make(chan interface{})
		out2 := make(chan interface{})
		go func() {
			defer close(out1)
			defer close(out2)
			for val := range orDone(done, in) {
				var out1, out2 = out1, out2
				for i := 0; i < 2; i++ {
					select {
					case out1 <- val:
						out1 = nil
					case out2 <- val:
						out2 = nil
					}
				}
			}
		}()

		return out1, out2
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

	out1, out2 := tee(done, test)

	for val1 := range out1 {
		fmt.Printf("out1: %v, out2:%v\n", val1, <-out2)
	}
}
