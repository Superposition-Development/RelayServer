package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

//

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

// needs to be sanitized
func QueryRow(returnValues []string, tableName string, columnToQuery string, inputValue string) (map[string]any, error) {
	formattedReturnValues := strings.Join(returnValues, ", ")

	db, err := sql.Open("sqlite", config.DatabaseName+".db")
	if err != nil {
		log.Fatalf("Couldn't open database: %v", err)
	}

	queryPrompt := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ?",
		formattedReturnValues,
		tableName,
		columnToQuery,
	)

	values := make([]any, len(returnValues))
	valuePtrs := make([]any, len(returnValues))

	row := db.QueryRow(queryPrompt, inputValue)

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = row.Scan(valuePtrs...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]any)
	for i, col := range returnValues {
		resultMap[col] = values[i]
	}

	return resultMap, nil
}
