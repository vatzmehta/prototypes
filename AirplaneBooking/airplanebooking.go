package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/vatzmehta/prototypes/utils"
)

/*
This code is to understand different types of SQL Locks.
We would go ahead with the following locks
1. For Update NoWait
2. For Update Skip Locked
3. For Update
*/

var db *sql.DB

type Users struct {
	ID   int
	Name string
}

type Flights struct {
	ID   int
	Name string
}

type Seats struct {
	ID         int
	FlightID   int
	UserID     sql.NullInt64
	SeatNumber string `db:"seat_number"`
}

func main() {
	db = utils.ConnectToSQLDB("airplane_booking")
	defer db.Close()
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(10 * time.Minute)

	tearDown(1)
	//printSeat(1)

	barrier := make(chan struct{})

	wg := sync.WaitGroup{}
	for i := 1; i < 163; i++ {
		wg.Add(1)
		go func(uid int) {
			<-barrier // wait until all goroutines are ready
			fmt.Println(time.Now().UnixNano())
			if err := bookSeat(context.Background(), uid, 1, &wg); err != nil {
				fmt.Printf("User %d failed to book a seat: %v\n", uid, err)
			}
		}(i)
	}

	barrier <- struct{}{} // signal all goroutines to start booking
	close(barrier)        // release all goroutines at once

	wg.Wait()
	printSeat(1)

}

// bookSeat books seat of a flight by making an entry in the seat table
func bookSeat(ctx context.Context, userID int, flightID int, wg *sync.WaitGroup) error {
	defer wg.Done()
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow("SELECT * FROM seats WHERE flight_id = ? AND user_id IS NULL LIMIT 1 FOR UPDATE", flightID)

	var seat Seats

	// fmt.Print(row)
	if err := row.Scan(&seat.ID, &seat.UserID, &seat.FlightID, &seat.SeatNumber); err != nil {
		tx.Rollback()
		fmt.Print(err)
		return err
	}

	if _, err := tx.Exec("UPDATE seats set user_id = ? WHERE id = ?", userID, seat.ID); err != nil {
		tx.Rollback()
		return err
	}
	//fmt.Print("here")
	return tx.Commit()
}

// printSeat queries all seats for a flight and prints their booking status as in an airplane
func printSeat(flightID int) {
	rows, err := db.Query("SELECT seat_number, user_id FROM seats WHERE flight_id = ? ORDER BY seat_number", flightID)
	if err != nil {
		println("Error querying seats:", err.Error())
		return
	}
	defer rows.Close()

	seatStatus := make(map[string]bool) // true if booked
	seatOrder := []string{}
	for rows.Next() {
		var seatNumber string
		var userID sql.NullInt64
		if err := rows.Scan(&seatNumber, &userID); err != nil {
			println("Error scanning seat:", err.Error())
			return
		}
		seatStatus[seatNumber] = userID.Valid
		seatOrder = append(seatOrder, seatNumber)
	}

	// Print all seats in order, 6 per row (A-F)
	for i, seat := range seatOrder {
		if seatStatus[seat] {
			print("X ")
		} else {
			print("- ")
		}
		if (i+1)%6 == 0 {
			println()
		}
	}
	if len(seatOrder)%6 != 0 {
		println()
	}
}

func tearDown(flightID int) {
	_, err := db.Exec("UPDATE seats SET user_id = NULL WHERE flight_id = ?", flightID)
	if err != nil {
		println("Error tearing down flight:", err.Error())
	}
}
