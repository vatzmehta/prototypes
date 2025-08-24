package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/*
Implement a simple connection pool in Go.
Connection pool persists connections to a databases and manages their lifecycle
They are helpful as a means to reduce the overhead of establishing connections to a database by reusing existing connections.
*/
func main() {

	db, err := sql.Open("mysql", "root:root@123@tcp(localhost:3306)/prototypes")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	start := time.Now()

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(10 * time.Minute)

	wg := &sync.WaitGroup{}
	wg.Add(200)

	for i := 0; i < 200; i++ {
		go testDatabaseConnection(db, wg)
	}

	wg.Add(200)
	for i := 0; i < 200; i++ {
		go testDatabaseConnection(db, wg)
	}

	wg.Add(1)
	go func() {
		var once sync.Once
		for {

			stats := db.Stats()
			fmt.Printf("Open: %d, InUse: %d, Idle: %d\n", stats.OpenConnections, stats.InUse, stats.Idle)
			time.Sleep(1 * time.Second)
			once.Do(func() {
				wg.Done()
			})
		}
	}()

	wg.Wait()
	fmt.Println("All goroutines finished in:", time.Since(start))
}

func testDatabaseConnection(db *sql.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	rows, err := db.Query("SELECT * FROM dummy")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			panic(err)
		}
		// fmt.Printf("id: %d, name: %s\n", id, name)
	}
}
