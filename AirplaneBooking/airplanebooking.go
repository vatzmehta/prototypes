package airplanebooking

import "github.com/vatzmehta/prototypes/utils"

/*
This code is to understand different types of SQL Locks.
We would go ahead with the following locks
1. For Update NoWait
2. For Update Skip Locked
3. For Update
*/
func main() {
	utils.ConnectToSQLDB("airplane_booking")
}
