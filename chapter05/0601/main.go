package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}

		orDone := make(chan interface{})
		go func() {
			defer close(orDone)
			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()

		return orDone
	}

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

	type startGoroutineFn func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) (heartbeat <-chan interface{})

	newSteward := func(timeout time.Duration, startGoroutine startGoroutineFn) startGoroutineFn {
		return func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
			heartbeat := make(chan interface{})
			go func() {
				defer close(heartbeat)

				var wardDone chan interface{}
				var wardHeartbeat <-chan interface{}
				startWard := func() {
					wardDone = make(chan interface{})
					wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
				}
				startWard()
				pulse := time.Tick(pulseInterval)

			monitorLoop:
				for {
					timeoutSignal := time.After(timeout)

					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct{}{}:
							default:
							}
						case <-wardHeartbeat:
							continue monitorLoop
						case <-timeoutSignal:
							log.Println("steward: ward unhealthy; restarting")
							close(wardDone)
							startWard()
							continue monitorLoop
						case <-done:
							return
						}
					}
				}
			}()

			return heartbeat
		}
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	doWorkFn := func(done <-chan interface{}, intList ...int) (startGoroutineFn, <-chan interface{}) {
		intChanStream := make(chan (<-chan interface{}))
		intStream := bridge(done, intChanStream)
		doWork := func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
			intChan := make(chan interface{})
			heartbeat := make(chan interface{})
			go func() {
				defer close(intChan)
				select {
				case intChanStream <- intChan:
				case <-done:
					return
				}

				pulse := time.Tick(pulseInterval)

				for {
				valueLoop:
					for _, intVal := range intList {
						if intVal < 0 {
							log.Printf("negative value: %v\n", intVal)
							return
						}

						for {
							select {
							case <-pulse:
								select {
								case heartbeat <- struct{}{}:
								default:
								}
							case intChan <- intVal:
								continue valueLoop
							case <-done:
								return
							}
						}
					}
				}
			}()

			return heartbeat
		}

		return doWork, intStream
	}

	done := make(chan interface{})
	defer close(done)

	doWork, intChan := doWorkFn(done, 1, 2, -1, 3, 4, 5)
	doWorkWithSteward := newSteward(1*time.Millisecond, doWork)
	doWorkWithSteward(done, 1*time.Hour)

	for intVal := range take(done, intChan, 6) {
		fmt.Printf("Received: %v\n", intVal)
	}
}
