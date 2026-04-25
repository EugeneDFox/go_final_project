package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

// db is the global database connection pool.
var db *sql.DB

// schema defines the SQLite database schema for the scheduler application.
// It creates a table for tasks with fields: id, date, title, comment, repeat.
// An index on the date column is also created for faster date-based queries.
const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
	title VARCHAR(256) NOT NULL DEFAULT "",
	comment TEXT NOT NULL DEFAULT "",
	repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX scheduler_date ON scheduler (date);
`

// Init initializes the database connection.
// It reads the TODO_DBFILE environment variable to determine the database file path
// (defaults to "scheduler.db" if not set). If the database file does not exist,
// it creates the file and executes the schema to set up the tables.
// Returns an error if the connection cannot be established or schema creation fails.
func Init() error {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}
	_, err := os.Stat(dbFile)
	var install bool
	if err != nil {
		install = true
	}
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	if install {
		_, err := db.Exec(schema)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the global database connection.
// It should be called when the application shuts down to release resources.
func Close() {
	db.Close()
}
