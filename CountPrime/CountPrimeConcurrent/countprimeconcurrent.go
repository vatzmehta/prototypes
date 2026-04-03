package main

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vatzmehta/prototypes/utils"
)

// 100 Million
const MAX_INT = 100000000

var PrimeCount int32

func checkPrime(n int, m *sync.Mutex) {

	// return if negative
	if n <= 1 {
		return
	}

	// for 2, increment count and written
	if n == 2 {
		m.Lock()
		PrimeCount++
		m.Unlock()
		return
	}

	// return if other even numbers
	if n&1 == 0 {
		return
	}

	for i := 3; i <= int(math.Sqrt(float64(n))); i += 2 {
		if n%i == 0 {
			return

		}
	}

	// m.Lock()
	// PrimeCount++
	// m.Unlock()
	atomic.AddInt32(&PrimeCount, 1)
}

func checkPrimeConcurrent(start, end int, m *sync.Mutex, wg *sync.WaitGroup) {
	timeStart := time.Now()

	for i := start; i < end; i++ {
		checkPrime(i, m)
	}

	fmt.Printf("Thread with range %d:%d took %v\n", start, end, time.Since(timeStart))
	wg.Done()
}

func main() {
	done := make(chan struct{})
	readings := utils.StartCPUMonitor(done)

	timeStart := time.Now()
	var m sync.Mutex
	wg := sync.WaitGroup{}
	for i := 0; i < MAX_INT; i += 10000000 {
		wg.Add(1)
		fmt.Printf("Checking for the range %d:%d\n", i, i+10000000)
		go checkPrimeConcurrent(i, i+10000000, &m, &wg)
	}

	wg.Wait()
	close(done)
	time.Sleep(10 * time.Millisecond)

	// Answer should be 5,761,455
	fmt.Println("Total Prime Count: ", PrimeCount)
	fmt.Println("Time taken: ", time.Since(timeStart).Seconds())
	utils.PrintCPUGraph(*readings)
}
