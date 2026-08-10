package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Tag struct {
	ID   int64
	Name string
}

type File struct {
	ID   int64
	Name string
}

var (
	dbHandle        *sql.DB
	dbPath          string
	dbName          string = "dj.sqlite"
	playlistDirName string = "dj_playlist"
)

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
	mux.HandleFunc("GET /api/media/{file_id}", MediaHandler)

	return mux
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "About page")
}

// Streaming a media file over http
func MediaHandler(w http.ResponseWriter, r *http.Request) {
	file_id := r.PathValue("file_id")

	var file File

	err := dbHandle.QueryRow(
		"SELECT id, name FROM file WHERE id = ?",
		file_id,
	).Scan(&file.ID, &file.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintln(w, "File not found")
			return
		}
		fmt.Fprintln(w, "Unknown error")
		return
	}

	path := filepath.Join(dbPath, file.Name)
	http.ServeFile(w, r, path)
}

// Closes the database connection and resets the dbHandle and dbPath global variables
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
	dbHandle, err = sql.Open("sqlite", filepath.Join(path, dbName)+"?_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}

	dbPath = path

	// Create the playlist directory
	err = os.MkdirAll(filepath.Join(path, playlistDirName), 0755)
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

// List files by tag IDs
// If the first tag ID is "0", all untagged files will be returned
// If tagIDs is empty, all files will be returned
func ListFiles(tagIDs []int64) ([]File, error) {
	var (
		files       []File
		queryString string
		queryArgs   []any
	)

	if len(tagIDs) == 0 {
		// Return all files
		queryString = `
			SELECT id, name
			FROM file
			ORDER BY id;
		`
	} else if tagIDs[0] == 0 {
		// Return untagged files
		queryString = `
			SELECT f.id, f.name
			FROM file AS f
			LEFT JOIN file_tag AS ft ON ft.file_id = f.id
			WHERE ft.file_id IS NULL
			ORDER BY f.id;
		`
	} else {
		// Return files by tagIDs
		placeholders := strings.TrimRight(strings.Repeat("?,", len(tagIDs)), ",")

		queryString = fmt.Sprintf(`
			SELECT f.id, f.name
			FROM file AS f
			JOIN file_tag AS ft ON ft.file_id = f.id
			WHERE ft.tag_id IN (%s)
			GROUP BY f.id, f.name
			HAVING COUNT(*) = ?
			ORDER BY f.id;
		`, placeholders)

		queryArgs = make([]any, len(tagIDs))
		for i, tagID := range tagIDs {
			queryArgs[i] = tagID
		}
		queryArgs = append(queryArgs, len(tagIDs))
	}

	// Execute the query
	rows, err := dbHandle.Query(queryString, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file File

		if err := rows.Scan(&file.ID, &file.Name); err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, rows.Err()
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
	rows, err := dbHandle.Query("SELECT id, name FROM tag ORDER BY id")
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

func DetachTag(fileID int, tagIDs []int) error {
	tx, err := dbHandle.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		DELETE FROM file_tag WHERE file_id=? AND tag_id=?;
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

// Creates a shuffled playlist by the provided tag groups and amounts
// returns the created file's absolute path
func CreatePlaylist(tagGroups []struct {
	TagsIDs []int64
	Amount  int
}) (string, error) {

	if len(tagGroups) == 0 {
		return "", fmt.Errorf("empty tag group array provided")
	}

	tx, err := dbHandle.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var files []File

	// IDs selected by earlier groups to prevent multiple selection of same files
	selectedIDs := make(map[int64]struct{})

	for _, tagGroup := range tagGroups {
		if len(tagGroup.TagsIDs) == 0 {
			return "", fmt.Errorf("tag group is empty")
		}

		if tagGroup.Amount <= 0 {
			return "", fmt.Errorf("tag group amount is 0")
		}

		// Looking for duplicate ids, and returning error if found
		tagSet := make(map[int64]struct{}, len(tagGroup.TagsIDs))

		for _, tagID := range tagGroup.TagsIDs {
			if _, exists := tagSet[tagID]; exists {
				return "", fmt.Errorf("duplicate tag ID %d in tag group", tagID)
			}

			tagSet[tagID] = struct{}{}
		}

		// For the (?, ?, ?, ...)
		tagPlaceholders := make([]string, len(tagGroup.TagsIDs))
		// Arguments to provide to the query in the end
		args := make([]any, 0, len(tagGroup.TagsIDs)+len(selectedIDs)+1)

		for i, tagID := range tagGroup.TagsIDs {
			tagPlaceholders[i] = "?"
			args = append(args, tagID)
		}

		query := `
			SELECT f.id, f.name
			FROM file AS f
			JOIN file_tag AS ft ON ft.file_id = f.id
			WHERE ft.tag_id IN (` + strings.Join(tagPlaceholders, ",") + `)
		`

		// Exclude files already selected by previous groups.
		if len(selectedIDs) > 0 {
			excludePlaceholders := make([]string, 0, len(selectedIDs))

			for id := range selectedIDs {
				excludePlaceholders = append(excludePlaceholders, "?")
				args = append(args, id)
			}

			query += `
				AND f.id NOT IN (` + strings.Join(excludePlaceholders, ",") + `)
			`
		}

		query += `
			GROUP BY f.id, f.name
			HAVING COUNT(DISTINCT ft.tag_id) = ?
			ORDER BY RANDOM()
			LIMIT ?
		`

		args = append(args, len(tagGroup.TagsIDs), tagGroup.Amount)

		rows, err := tx.Query(query, args...)
		if err != nil {
			return "", err
		}

		for rows.Next() {
			var f File

			if err := rows.Scan(&f.ID, &f.Name); err != nil {
				rows.Close()
				return "", err
			}

			files = append(files, f)
			selectedIDs[f.ID] = struct{}{}
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			return "", err
		}

		rows.Close()

	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	// Shuffling the files
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fileNames := make([]string, len(files))

	for i, file := range files {
		fileNames[i] = file.Name
	}

	r.Shuffle(len(fileNames), func(i, j int) {
		fileNames[i], fileNames[j] = fileNames[j], fileNames[i]
	})

	playlistPath, err := CreateM3U8(fileNames)

	if err != nil {
		return "", err
	}

	return playlistPath, nil

}

// Creates a playlist file and returns the created file's absolute path
func CreateM3U8(mediaFiles []string) (string, error) {
	filename := fmt.Sprintf(
		"dj-%s.m3u8",
		time.Now().Format("2006-01-02-15-04-05"),
	)

	outputPath := filepath.Join(dbPath, playlistDirName, filename)

	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	if _, err := writer.WriteString("#EXTM3U\n"); err != nil {
		return "", err
	}

	for _, mediaFile := range mediaFiles {
		if _, err := writer.WriteString(mediaFile + "\n"); err != nil {
			return "", err
		}
	}

	if err := writer.Flush(); err != nil {
		return "", err
	}

	return outputPath, nil
}
