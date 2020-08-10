package main

import "fmt"

func main() {
	stringCh := make(chan string)
	go func() {
		stringCh <- "Hello channels!"
	}()
	fmt.Println(<-stringCh)
}
