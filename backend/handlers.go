package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

func SetupServer() http.Handler {
	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("POST /choose-dir", ChooseDirHandler)
	mux.HandleFunc("GET /tags", ListTagsHandler)
	mux.HandleFunc("GET /tags/{file_id}", TagsByFileIDHandler)
	mux.HandleFunc("GET /media/{file_id}", MediaHandler)

	// Middlewares
	// Applied from innermost to outermost.
	handler := maxBodySizeMiddleware(mux)
	// handler = anotherMiddleware(handler)

	return handler
}

// To protect against huge bodies.
func maxBodySizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
		defer r.Body.Close()

		next.ServeHTTP(w, r)
	})
}

// Writes JSON into response
func writeResJSON(w http.ResponseWriter, status int, data map[string]any) {
	resBody, err := json.Marshal(data)
	if err != nil {
		// data could not be marshalled
		log.Printf("failed to marshal JSON response: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(resBody); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

// Choose a directory and set up the database in the backend.
func ChooseDirHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DirPath string `json:"dirPath"`
	}

	// Invalid JSON
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid JSON",
		})
		return
	}

	// dirPath missing
	if req.DirPath == "" {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "dirPath is required",
		})
		return
	}

	// Initialize database
	if err := InitDatabase(req.DirPath); err != nil {
		log.Println(err)

		writeErr := func(err error) {
			writeResJSON(w, http.StatusInternalServerError, map[string]any{
				"error": err,
			})
		}

		switch {
		case errors.Is(err, ErrOpeningDatabase):
			writeErr(ErrOpeningDatabase)

		case errors.Is(err, ErrDatabaseConnection):
			writeErr(ErrDatabaseConnection)

		case errors.Is(err, ErrPlaylistDirCreate):
			writeErr(ErrPlaylistDirCreate)

		case errors.Is(err, ErrTableCreation):
			writeErr(ErrTableCreation)

		case errors.Is(err, ErrUpdatingFiles):
			writeErr(ErrUpdatingFiles)

		default:
			writeResJSON(w, http.StatusInternalServerError, map[string]any{
				"error": "Unknown error",
			})
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := ListTagsAll()
	if err != nil {
		log.Println(err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to query database",
		})

		return
	}

	writeResJSON(w, http.StatusOK, map[string]any{
		"tags": tags,
	})
}

func TagsByFileIDHandler(w http.ResponseWriter, r *http.Request) {
	param_file_id := r.PathValue("file_id")

	// Validating file ID
	file_id, err := strconv.ParseInt(param_file_id, 10, 64)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid file id",
		})
		return
	}
	if file_id <= 0 {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "file id has to be bigger than 0",
		})
		return
	}

	// Listing Tags
	tags, err := ListTagsByFileID(file_id)
	if err != nil {
		log.Println(err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to query database",
		})
		return
	}

	writeResJSON(w, http.StatusOK, map[string]any{
		"tags": tags,
	})
}

// Streaming a media file over http
func MediaHandler(w http.ResponseWriter, r *http.Request) {
	param_file_id := r.PathValue("file_id")

	// Validating file ID
	file_id, err := strconv.ParseInt(param_file_id, 10, 64)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid file id",
		})
		return
	}
	if file_id <= 0 {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "file id has to be bigger than 0",
		})
		return
	}

	var file File

	err = dbHandle.QueryRow(
		"SELECT id, name FROM file WHERE id = ?",
		file_id,
	).Scan(&file.ID, &file.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			writeResJSON(w, http.StatusNotFound, map[string]any{
				"error": "File not found",
			})
			return
		} else {
			writeResJSON(w, http.StatusInternalServerError, map[string]any{
				"error": "Failed to stream file",
			})
			return
		}
	}

	path := filepath.Join(dbPath, file.Name)
	http.ServeFile(w, r, path)
}
