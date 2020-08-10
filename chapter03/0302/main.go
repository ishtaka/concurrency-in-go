package main

import "fmt"

func main() {
	stringCh := make(chan string)
	go func() {
		stringCh <- "Hello channels!"
	}()
	salutation, ok := <-stringCh
	fmt.Printf("(%v): %v", ok, salutation)
}
