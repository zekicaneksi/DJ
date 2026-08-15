package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

var (
	dbHandle *sql.DB
	dbPath   string
)

// Closes the database connection and resets the dbHandle and dbPath global variables
func CloseDB() error {
	if dbHandle != nil {
		if err := dbHandle.Close(); err != nil {
			return fmt.Errorf("error when closing database: %w", err)
		}
		dbHandle = nil
		dbPath = ""
	}

	return nil
}

// Creates/opens the database file and creates the tables if do not exist
// Sets the global dbPath and dbHandle variables
// Checks the files in path and updates the database
func InitDatabase(path string) error {
	// Close connection if there is one already
	if err := CloseDB(); err != nil {
		return err
	}

	// Create/Open the database file
	var err error
	dbHandle, err = sql.Open("sqlite", filepath.Join(path, dbName)+"?_pragma=foreign_keys(1)")
	if err != nil {
		return fmt.Errorf("error when opening database: %w", err)
	}

	// Check connection to the database
	if err := dbHandle.Ping(); err != nil {
		return fmt.Errorf("error when pinging database: %w", err)
	}

	dbPath = path

	// Create the playlist directory
	err = os.MkdirAll(filepath.Join(path, playlistDirName), 0755)
	if err != nil {
		return fmt.Errorf("error when creating the playlist directory: %w", err)
	}

	// Create database tables if do not exist
	_, err = dbHandle.Exec(`
		CREATE TABLE IF NOT EXISTS file (
		id INTEGER PRIMARY KEY NOT NULL,
    	name TEXT NOT NULL UNIQUE
        	CHECK(length(trim(name)) > 0)
        	CHECK(length(name) <= 100)
		);

		CREATE TABLE IF NOT EXISTS tag (
		id INTEGER PRIMARY KEY NOT NULL,
    	name TEXT NOT NULL UNIQUE
        	CHECK(length(trim(name)) > 0)
        	CHECK(length(name) <= 50)
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
		return fmt.Errorf("error when creating database tables: %w", err)
	}

	if err := UpdateFiles(); err != nil {
		return fmt.Errorf("error when updating files: %w", err)
	}

	return nil
}
