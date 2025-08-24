package utils

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectToSQLDB(dbName string) *sql.DB {
	db, err := sql.Open("mysql", "root:root@123@tcp(localhost:3306)/"+dbName)
	if err != nil {
		panic(err)
	}
	return db
}
