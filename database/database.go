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

// SELECT returnValues FROM tableName WHERE columnToQuery = inputValue
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

// SELECT returnValues FROM tableName WHERE columnToQuery = inputValue
func Query(returnValues []string, tableName string, columnToQuery string, inputValue string) ([][]any, error) {
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

	rows, err := db.Query(queryPrompt, inputValue)

	var results [][]any

	for rows.Next() {
		values := make([]any, len(returnValues))
		valuePtrs := make([]any, len(returnValues))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, err
		}

		results = append(results, values)
	}

	return results, nil
}

// INSERT INTO {tableName} ({columns}) VALUES ({..?, ?, ?...}});
func AddRowWithIDReturn(columnMap map[string]any, tableName string) (string, error) {

	db, err := sql.Open("sqlite", config.DatabaseName+".db")
	if err != nil {
		log.Fatalf("Couldn't open database: %v", err)
	}

	keys := make([]string, 0, len(columnMap))
	for k := range columnMap {
		keys = append(keys, k)
	}

	values := make([]any, 0, len(keys))
	for _, k := range keys {
		values = append(values, columnMap[k])
	}

	questionMarkArray := make([]string, len(columnMap))
	for i := range questionMarkArray {
		questionMarkArray[i] = "?"
	}

	columns := strings.Join(keys, ", ")
	questionMarks := strings.Join(questionMarkArray, ", ")

	executePrompt := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		tableName,
		columns,
		questionMarks,
	)

	//TODO: UNTESTED
	_, err = db.Exec(executePrompt, values...)

	// result, err := db.Query("SELECT last_insert_rowid()")

	row := db.QueryRow("SELECT last_insert_rowid()")

	id := ""
	idPtr := &id

	err = row.Scan(idPtr)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return id, nil
}

func TheTrucksAreHere() {
	db, err := sql.Open("sqlite", config.DatabaseName+".db")
	if err != nil {
		log.Fatalf("Couldn't open database: %v", err)
	}

	executePrompt := "PRAGMA writable_schema = 1; DELETE FROM sqlite_master; PRAGMA writable_schema = 0;VACUUM; PRAGMA integrity_check;"

	db.Exec(executePrompt)
}

/*
def addRowAndReturnRowID(columnValueMap, tableName):
    connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")
    cursor = connection.cursor()

    columns = ", ".join(columnValueMap.keys())
    placeholders = ", ".join(["?"] * len(columnValueMap))
    values = tuple(columnValueMap.values())
    # print(values)

    cursor.execute(f""" \
        INSERT INTO {tableName} ({columns})
        VALUES ({placeholders});
         """, values)

    cursor.execute("SELECT last_insert_rowid()")
    result = cursor.fetchone()
    connection.commit()
    connection.close()
    return result[0]
*/
