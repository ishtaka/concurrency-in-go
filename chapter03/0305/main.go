package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	var stdoutBuff bytes.Buffer
	defer func() { _, _ = stdoutBuff.WriteTo(os.Stdout) }()

	intCh := make(chan int, 4)
	go func() {
		defer close(intCh)
		defer func() { _, _ = fmt.Fprintln(&stdoutBuff, "Producer Done.") }()
		for i := 0; i < 5; i++ {
			_, _ = fmt.Fprintf(&stdoutBuff, "Sending: %d\n", i)
			intCh <- i
		}
	}()

	for integer := range intCh {
		_, _ = fmt.Fprintf(&stdoutBuff, "Received %v.\n", integer)
	}
}
