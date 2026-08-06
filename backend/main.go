package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

var dbHandle *sql.DB

func main() {
	// Close the database before exiting app
	closeDB := func() {
		if err := CloseDB(); err != nil {
			fmt.Println(err)
		}
	}
	defer closeDB()

	// Set up server
	mux := SetupServer()

	// Listen
	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func SetupServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", HomeHandler)
	mux.HandleFunc("GET /about", AboutHandler)

	return mux
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "About page")
}

func CloseDB() error {
	if dbHandle != nil {
		if err := dbHandle.Close(); err != nil {
			return err
		}
	}

	return nil
}

func InitDatabase(path string) error {
	// Close connection if there is one already
	CloseDB()

	// Create/Open the database file
	dbHandle, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	// Check connection to the database
	if err := dbHandle.Ping(); err != nil {
		return err
	}

	// Create database tables if do not exist
	_, err = dbHandle.Exec(`
		CREATE TABLE IF NOT EXISTS file (
		id INTEGER PRIMARY KEY NOT NULL,
		name varchar(255) UNIQUE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS tag (
		id INTEGER PRIMARY KEY NOT NULL,
		name varchar(255) UNIQUE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS file_tag (
		file_id INT NOT NULL,
		tag_id INT NOT NULL,
		FOREIGN KEY(file_id) REFERENCES file(id) ON DELETE CASCADE,
		FOREIGN KEY(tag_id) REFERENCES tag(id) ON DELETE CASCADE,
		UNIQUE(file_id, tag_id)
		);
	`)
	if err != nil {
		return err
	}

	return nil
}
