package main

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vatzmehta/prototypes/utils"
)

var PrimeCount int32
var CurrentCounter int32

const MAX_INT = 100000000

func checkPrime(n int) {

	if n <= 1 {
		return
	}

	if n == 2 {
		atomic.AddInt32(&PrimeCount, 1)
		return
	}

	// return even numbers
	if n&1 == 0 {
		return
	}

	for i := 3; i <= int(math.Sqrt(float64(n))); i += 2 {
		if n%i == 0 {
			return

		}
	}

	atomic.AddInt32(&PrimeCount, 1)
}

func doWork(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		x := atomic.AddInt32(&CurrentCounter, 1)
		if x > MAX_INT {
			break
		}

		checkPrime(int(x))
	}

}

func main() {
	done := make(chan struct{})
	readings := utils.StartCPUMonitor(done)

	timeStart := time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go doWork(&wg)
	}

	wg.Wait()
	close(done)
	time.Sleep(10 * time.Millisecond)

	// Answer should be 5,761,455
	fmt.Println("Total Prime Count: ", PrimeCount)
	fmt.Println("Time taken: ", time.Since(timeStart).Seconds())
	utils.PrintCPUGraph(*readings)
}
