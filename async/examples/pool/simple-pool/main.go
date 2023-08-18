package main

import (
	"context"
	"fmt"
	"time"

	"github.com/OrenRosen/go/async"
)

func main() {
	// initialize the pool
	// by default it opens 10 workers, each worker listens in a different go routine on a channel for a received function
	pool := async.NewPool()

	// call `pool.RunAsync` with a context and a closure.
	// this will add the passed function to the queue channel for be consumed by an available worker
	pool.RunAsync(context.Background(), func(ctx context.Context) error {
		fmt.Println("running in async pool")
		return nil
	})

	// for the example, sleeping in order to see the print from the async function
	fmt.Println("going to sleep...")
	time.Sleep(time.Second)
}
