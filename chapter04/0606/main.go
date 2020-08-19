package main

import "fmt"

func main() {
	repeat := func(done <-chan interface{}, values ...interface{}) <-chan interface{} {
		valueChan := make(chan interface{})
		go func() {
			defer close(valueChan)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueChan <- v:
					}
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
	toString := func(done <-chan interface{}, valueChan <-chan interface{}) <-chan string {
		stringChan := make(chan string)
		go func() {
			defer close(stringChan)
			for v := range valueChan {
				select {
				case <-done:
					return
				case stringChan <- v.(string):
				}
			}
		}()

		return stringChan
	}

	done := make(chan interface{})
	defer close(done)

	var message string
	for token := range toString(done, take(done, repeat(done, "I", "am."), 5)) {
		message += token
	}

	fmt.Printf("message: %s", message)
}
