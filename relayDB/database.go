package relayDB

import (
	_ "RelayServer/relayConfig"

	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)


// FIXME: (If possible) Paths are like this since they're relative to main.go

func InitializeDB() {
	sqlByte, err := os.ReadFile("../relayDB/init.sql")
	if err != nil {
		log.Fatalf("Couldn't read init.sql: %v", err)
	}
	sqlScript := string(sqlByte)

	// Jon i have no idea where you're getting that config var from :(
	db, err := sql.Open("sqlite", "../relayDB/d9xb1.db")
	if err != nil {
		log.Fatalf("Couldn't open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("Couldn't enable foreign keys: %v", err)
	}

	_, err = db.Exec(sqlScript)
	if err != nil {
		log.Fatalf("Failed to execute SQL script: %v", err)
	}

	fmt.Println("Relay DB Initialized")
}
