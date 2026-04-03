package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// 100 Million
const MAX_INT = 100000000

var PrimeCount = 0

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

	m.Lock()
	PrimeCount++
	m.Unlock()
}

func checkPrimeConcurrent(start, end int, m *sync.Mutex, wg *sync.WaitGroup) {

	for i := start; i < end; i++ {
		checkPrime(i, m)
	}

	wg.Done()
}

func main() {
	timeStart := time.Now()
	var m sync.Mutex
	wg := sync.WaitGroup{}
	for i := 0; i < MAX_INT; i += 10000000 {
		wg.Add(1)
		fmt.Printf("Checking for the range %d:%d\n", i, i+10000000)
		go checkPrimeConcurrent(i, i+10000000, &m, &wg)
	}

	wg.Wait()
	println("Total Prime Count: ", PrimeCount)
	println("Time taken: ", time.Since(timeStart).Seconds())
}
