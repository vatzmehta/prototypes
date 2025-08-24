package main

import (
	"sync"
	"time"
)

var channel chan interface{}

// Push to the blocking queue.
// Waits ifthe queue is full.
func Push(i interface{}) {
	channel <- i
}

// Pop from the blocking queue.
// Waits if the queue is empty.
func Pop(i *interface{}) {
	*i = <-channel
}

func main() {
	channel = make(chan interface{}, 3)

	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			Push(i)
			println("Pushed:", i)
		}(i)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var v interface{}
			now := time.Now()
			Pop(&v)
			println("Popped after", time.Since(now), ":", v.(int))
		}()
	}

	wg.Wait()
	close(channel) // Close the channel to prevent further sends.

}
