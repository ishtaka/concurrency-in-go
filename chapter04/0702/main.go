package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

func main() {
	repeatFn := func(done <-chan interface{}, fn func() interface{}) <-chan interface{} {
		valueChan := make(chan interface{})
		go func() {
			defer close(valueChan)
			for {
				select {
				case <-done:
					return
				case valueChan <- fn():
				}
			}
		}()

		return valueChan
	}

	take := func(done <-chan interface{}, valueChan <-chan interface{}, num int) <-chan interface{} {
		takeChan := make(chan interface{})
		go func() {
			defer close(takeChan)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeChan <- <-valueChan:
				}
			}
		}()

		return takeChan
	}

	toInt := func(done <-chan interface{}, valueChan <-chan interface{}) <-chan int {
		intChan := make(chan int)
		go func() {
			defer close(intChan)
			for v := range valueChan {
				select {
				case <-done:
					return
				case intChan <- v.(int):
				}
			}
		}()

		return intChan
	}

	primeFinder := func(done <-chan interface{}, intChan <-chan int) <-chan interface{} {
		isPrime := func(n int) bool {
			for i := 2; i < n; i++ {
				if n%i == 0 {
					return false
				}
			}

			return n > 1
		}

		primeChan := make(chan interface{})
		go func() {
			defer close(primeChan)
			for n := range intChan {
				if isPrime(n) {
					select {
					case <-done:
						return
					case primeChan <- n:
					}
				} else {
					select {
					case <-done:
						return
					default:
					}
				}
			}
		}()

		return primeChan
	}

	fanIn := func(done <-chan interface{}, channels ...<-chan interface{}) <-chan interface{} {
		var wg sync.WaitGroup
		multiplexedChan := make(chan interface{})

		multiplex := func(c <-chan interface{}) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedChan <- i:
				}
			}
		}

		wg.Add(len(channels))
		for _, c := range channels {
			go multiplex(c)
		}

		go func() {
			wg.Wait()
			close(multiplexedChan)
		}()

		return multiplexedChan
	}

	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	random := func() interface{} { return rand.Intn(50000000) }
	randIntChan := toInt(done, repeatFn(done, random))

	numFinders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", numFinders)
	finders := make([]<-chan interface{}, numFinders)
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntChan)
	}

	fmt.Println("Primes:")
	for prime := range take(done, fanIn(done, finders...), 10) {
		fmt.Printf("\t%d\n", prime)
	}

	fmt.Printf("Search took: %v\n", time.Since(start))
}
