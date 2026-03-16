package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

func InitializeDB() {
	sqlByte, err := os.ReadFile("init.sql")
	if err != nil {
		log.Fatalf("Couldn't read init.sql: %v", err)
	}
	sqlScript := string(sqlByte)

	fmt.Println(config.DatabaseName)

	db, err := sql.Open("sqlite", config.DatabaseName+".db")
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

func QueryRow(returnValues []string) {
	formattedReturnValues := strings.Join(returnValues, ", ")

	db, err := sql.Open("sqlite", config.DatabaseName+".db")
	if err != nil {
		log.Fatalf("Couldn't open database: %v", err)
	}

	queryPrompt := "SELECT" + formattedReturnValues

	db.QueryRow(queryPrompt)
	defer db.Close()
}
