package main

import (
	"log"
	"os"

	"github.com/EugeneDFox/go_final_project/pkg/db"
	"github.com/EugeneDFox/go_final_project/pkg/server"
)

// main is the entry point of the task scheduler application.
// It initializes the database, creates a logger, starts the HTTP server,
// and runs until the server stops (or an error occurs).
func main() {
	// Initialize database connection
	err := db.Init()
	if err != nil {
		log.Fatal(err)
	}
	// Ensure database connection is closed when the program exits
	defer db.Close()

	// Create a logger with prefix "SERVER: " that includes file and line number
	logger := log.New(os.Stdout, "SERVER: ", log.LstdFlags|log.Lshortfile)

	// Create a new HTTP server instance with the logger
	s := server.NewServer(logger)

	// Start the server and block until it stops (or fails)
	log.Fatal(s.Start())
}
