package main

import (
	"context"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func Open() *APIConnection {
	return &APIConnection{
		apiLimit: MultiLimiter(
			rate.NewLimiter(Per(2, time.Second), 1),
			rate.NewLimiter(Per(10, time.Minute), 10),
		),
		diskLimit: MultiLimiter(
			rate.NewLimiter(rate.Limit(1), 1),
		),
		networkLimit: MultiLimiter(
			rate.NewLimiter(Per(3, time.Second), 3),
		),
	}
}

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

type APIConnection struct {
	networkLimit,
	diskLimit,
	apiLimit RateLimiter
}

func (con *APIConnection) ReadFile(ctx context.Context) error {
	if err := MultiLimiter(con.apiLimit, con.diskLimit).Wait(ctx); err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

func (con *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := MultiLimiter(con.apiLimit, con.networkLimit).Wait(ctx); err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

type RateLimiter interface {
	Wait(ctx context.Context) error
	Limit() rate.Limit
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{limiters: limiters}
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

func main() {
	defer log.Printf("Done.")
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	apiConnection := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile: %v", err)
			}
			log.Printf("ReadFile")
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ResolveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress: %v", err)
			}
			log.Printf("ResolveAddress")
		}()
	}

	wg.Wait()
}
