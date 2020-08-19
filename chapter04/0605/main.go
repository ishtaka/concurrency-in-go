package main

import (
	"fmt"
	"math/rand"
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

	done := make(chan interface{})
	defer close(done)

	random := func() interface{} { return rand.Int() }

	for num := range take(done, repeatFn(done, random), 10) {
		fmt.Println(num)
	}
}
