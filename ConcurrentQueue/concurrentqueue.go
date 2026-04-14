package main

import (
	"fmt"
	"sync"
)

type ConcurrentQueue struct {
	queue []int32
	mutex sync.Mutex
}

func (q *ConcurrentQueue) Enqueue(item int32) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.queue = append(q.queue, item)
}

func (q *ConcurrentQueue) Dequeue() int32 {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	// cannot call q.Size() here: it tries to acquire q.mutex, which this goroutine already holds (sync.Mutex is not re-entrant)
	if len(q.queue) == 0 {
		panic("Dequeue opeartion on empty queue")
	}

	item := q.queue[0]
	q.queue = q.queue[1:]
	return item
}

func (q *ConcurrentQueue) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return len(q.queue)
}

func main() {

	q := ConcurrentQueue{
		queue: make([]int32, 0),
	}

	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)

	var wgE sync.WaitGroup
	var wgD sync.WaitGroup

	for i := 0; i < 10; i++ {
		wgE.Add(1)
		go func() {
			cond.L.Lock()
			q.Enqueue(int32(i))
			cond.Signal()
			cond.L.Unlock()
			wgE.Done()
		}()
	}

	for i := 0; i < 10; i++ {

		wgD.Add(1)
		go func() {
			// Consumer                          Producer
			// --------                          --------
			// cond.L.Lock()
			// q.Size() == 0  ✓
			// cond.Wait()
			//   └─ Unlock()                     cond.L.Lock()   ← blocks until Wait releases
			//                                   q.Enqueue(item)
			//                                   cond.Signal()
			//                                   cond.L.Unlock()
			//   └─ wake up
			//   └─ Lock()    ← re-acquires
			// safe to Dequeue

			cond.L.Lock()
			// Use for instead of if: a Signal() does not guarantee the condition is still
			// true when this goroutine re-acquires the lock. Another dequeue goroutine may
			// have consumed the item first, so re-check before calling Dequeue.
			for q.Size() == 0 {
				fmt.Println("waiting")
				cond.Wait()
			}

			fmt.Println(q.Dequeue())
			cond.L.Unlock()
			wgD.Done()
		}()
	}

	wgE.Wait()
	wgD.Wait()
	fmt.Println("Size:", q.Size())

}
