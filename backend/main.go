package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type Tag struct {
	ID   int64
	Name string
}

var dbHandle *sql.DB
var dbPath string

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

	dbPath = ""

	return nil
}

// Creates/opens the database file and creates the tables if do not exist
// Sets the global dbPath and dbHandle variables
// Checks the files in path and updates the database
func InitDatabase(path string) error {
	// Close connection if there is one already
	CloseDB()

	// Create/Open the database file
	var err error
	dbHandle, err = sql.Open("sqlite", path+"/dj.sqlite?_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}

	dbPath = path

	// Check connection to the database
	if err := dbHandle.Ping(); err != nil {
		return err
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
        	CHECK(length(name) <= 255)
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

	if err := UpdateFiles(); err != nil {
		return err
	}

	return nil
}

// Checks the directory the database is in
// Adds new files into the database and removes the missing files from the database
func UpdateFiles() error {
	// -- Get files
	files, err := os.ReadDir(dbPath)
	if err != nil {
		return err
	}

	// Allowed media file extensions
	fileExtensions := []string{
		"mp4", "mkv", "avi", "mov", "wmv",
		"flv", "webm", "mp3", "aac", "flac", "wav",
	}

	// Making a map for faster look-up
	fileExtensionsMap := make(map[string]struct{}, len(fileExtensions))
	for _, ext := range fileExtensions {
		fileExtensionsMap[ext] = struct{}{}
	}

	// Filter the files from directories and non-media files
	var fileNames []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Name())), ".")

		if _, ok := fileExtensionsMap[ext]; !ok {
			continue
		}

		fileNames = append(fileNames, file.Name())
	}

	// -- Update the database
	tx, err := dbHandle.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Add new files
	stmt, err := tx.Prepare(`
		INSERT INTO file(name)
		VALUES(?)
		ON CONFLICT(name) DO NOTHING;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, name := range fileNames {
		if _, err := stmt.Exec(name); err != nil {
			return err
		}
	}

	// Remove missing files
	if len(fileNames) == 0 {
		_, err := tx.Exec("DELETE FROM file")
		if err != nil {
			return err
		}
	} else {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(fileNames)), ",")

		query := fmt.Sprintf(
			"DELETE FROM file WHERE name NOT IN (%s)",
			placeholders,
		)

		args := make([]any, len(fileNames))
		for i, n := range fileNames {
			args[i] = n
		}

		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil

}

// Creates a tag and returns the created tag's id
func CreateTag(tagName string) (int64, error) {
	result, err := dbHandle.Exec(
		"INSERT INTO tag (name) VALUES (?)",
		tagName,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Returns all of the tags at once
func ListTags() ([]Tag, error) {
	rows, err := dbHandle.Query("SELECT id, name FROM tag")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag

	for rows.Next() {
		var tag Tag

		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

func DeleteTag(tagID int64) error {
	_, err := dbHandle.Exec(
		"DELETE FROM tag WHERE id = ?",
		tagID,
	)
	if err != nil {
		return err
	}

	return nil
}

func RenameTag(tagID int64, newName string) error {
	_, err := dbHandle.Exec(`
		UPDATE tag
		SET name = ?
		WHERE id = ?
	`, newName, tagID)
	if err != nil {
		return err
	}

	return nil
}

// Attaches tags to a file
func AttachTag(fileID int, tagIDs []int) error {
	tx, err := dbHandle.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO file_tag (file_id, tag_id)
		VALUES(?,?);
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tagID := range tagIDs {
		if _, err := stmt.Exec(fileID, tagID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
