package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
)

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
