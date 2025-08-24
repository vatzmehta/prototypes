package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

type Dummy struct {
	Id   int
	Name string
}

type ResourceLock struct {
	mu      sync.Mutex
	condMap map[int]*ResourceState
	dataMap map[int]*Dummy
}

type ResourceState struct {
	cond     *sync.Cond
	fetching bool
}

func NewResourceLock() *ResourceLock {
	return &ResourceLock{
		condMap: make(map[int]*ResourceState),
		dataMap: make(map[int]*Dummy),
	}
}

func (rl *ResourceLock) GetResource(ctx context.Context, id int, client *redis.Client,
	db *sql.DB) *Dummy {
	fmt.Println("GetResource:", id)
	rl.mu.Lock()

	res, exist := rl.dataMap[id]
	if exist {
		rl.mu.Unlock()
		fmt.Print("Found in dataMap")
		return res
	}

	_, exist = rl.condMap[id]
	if !exist {
		rl.condMap[id] = &ResourceState{
			cond:     sync.NewCond(&rl.mu),
			fetching: true,
		}
		rl.mu.Unlock()

		d, _ := fetchFromDb(ctx, db, id)

		rl.dataMap[id] = d
		rl.condMap[id].fetching = false

		cacheKey := fmt.Sprintf("dummy:%d", id)
		fmt.Println("Setting cache for", id)
		data, _ := json.Marshal(d)
		client.Set(ctx, cacheKey, data, 5*time.Minute)
		fmt.Println("Broadcasting")
		rl.condMap[id].cond.Broadcast()
		delete(rl.condMap, id)
		fmt.Println("Broadcasted")
		delete(rl.dataMap, id)
		fmt.Print("queried and found from DB")
		return d
	}

	fmt.Println("Waiting for resource to be fetched")
	rl.condMap[id].cond.Wait()
	fmt.Println("wait fin")
	if res, exist = rl.dataMap[id]; exist {
		rl.mu.Unlock()
		fmt.Print("Found in dataMap - 2")
		return res
	}
	rl.mu.Unlock()
	cacheKey := fmt.Sprintf("dummy:%d", id)

	val, _ := client.Get(ctx, cacheKey).Result()
	fmt.Println("Cache hit - 2")
	var dummy1 Dummy
	if err := json.Unmarshal([]byte(val), &dummy1); err != nil {
		// ignoring err
	}
	return &dummy1
}

/*
Simple code to demonstrate cache stampede
*/
func main() {
	ctx := context.Background()
	rl := NewResourceLock()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	//testRedis(ctx, client)

	db, err := sql.Open("mysql", "root:root@123@tcp(localhost:3306)/prototypes")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(10 * time.Minute)

	wg := sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			d, _ := fetchFromDbWithCache(ctx, db, client, 1, rl)
			fmt.Println(d)
		}(i)
	}

	wg.Wait()

}

func fetchFromDbWithCache(ctx context.Context, db *sql.DB, client *redis.Client,
	id int, rl *ResourceLock) (*Dummy, error) {
	cacheKey := fmt.Sprintf("dummy:%d", id)

	val, err := client.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			// Cache miss,fetch from DB
			fmt.Printf("Cache miss, fetching from DB for %d\n", id)
			return rl.GetResource(ctx, id, client, db), nil
		} else {
			return nil, fmt.Errorf("error fetching from cache: %v", err)
		}
	} else {
		fmt.Println("Cache hit")
		var dummy1 Dummy
		if err := json.Unmarshal([]byte(val), &dummy1); err != nil {
			// ignoring err
		}
		return &dummy1, nil
	}

}

func fetchFromDBAndSetCache(ctx context.Context, db *sql.DB, client *redis.Client, id int) *Dummy {
	dummy1, _ := fetchFromDb(ctx, db, id)
	// persist in cache
	data, _ := json.Marshal(dummy1)
	cacheKey := fmt.Sprintf("dummy:%d", id)
	client.Set(ctx, cacheKey, data, 5*time.Minute)
	return dummy1
}

func fetchFromDb(ctx context.Context, db *sql.DB, id int) (*Dummy, error) {
	row, err := db.QueryContext(ctx, "SELECT * FROM dummy WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	for row.Next() {
		var id int
		var name string
		if err := row.Scan(&id, &name); err != nil {
			panic(err)
		}
		fmt.Printf("Fetched from DB: %d, %s\n", id, name)
		return &Dummy{Id: id, Name: name}, nil
	}

	return nil, nil
}

// testRedis tests redis by doing a ping and setting and getting
func testRedis(ctx context.Context, client *redis.Client) {
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(pong)

	client.Set(ctx, "mykey", "myvalue", 5*time.Second)

	time.Sleep(2 * time.Second)

	val, err := client.Get(ctx, "mykey").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(val)
}
