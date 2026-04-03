package main

import (
	"math"
	"time"
)

// 100 Million
const MAX_INT = 100000000

var PrimeCount = 1

func checkPrime(n int) {
	// return even numbers
	if n&1 == 0 {
		return
	}

	for i := 3; i <= int(math.Sqrt(float64(n))); i += 2 {
		if n%i == 0 {
			return

		}
	}

	PrimeCount++
}

func main() {
	timeStart := time.Now()
	for i := 2; i < MAX_INT; i++ {
		checkPrime(i)
	}

	// Answer should be 5,761,455
	println("Total Prime Count: ", PrimeCount)
	println("Time taken: ", time.Since(timeStart).Seconds())
}
