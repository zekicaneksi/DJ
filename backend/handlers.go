package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"
)

// Setup Server
func SetupServer() http.Handler {
	mux := http.NewServeMux()

	// Handlers
	mux.HandleFunc("POST /choose-dir", ChooseDirHandler)
	mux.HandleFunc("GET /tags", ListTagsHandler)
	mux.HandleFunc("GET /tags/{file_id}", TagsByFileIDHandler)
	mux.HandleFunc("POST /create-tag", CreateTagHandler)
	mux.HandleFunc("POST /rename-tag", RenameTagHandler)
	mux.HandleFunc("GET /media/{file_id}", MediaHandler)

	// Middlewares
	// Applied from innermost to outermost.
	handler := maxBodySizeMiddleware(mux)
	// handler = anotherMiddleware(handler)

	return handler
}

// To protect against huge bodies
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
		log.Printf("failed to marshal JSON %v: %v", data, err)
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
		log.Printf("failed to write JSON %v: %v", resBody, err)
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

	// dirPath missing or empty
	if req.DirPath == "" {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "dirPath is required",
		})
		return
	}

	// Initialize database
	if err := InitDatabase(req.DirPath); err != nil {
		log.Printf("failed to initialize db for path %s: %v", req.DirPath, err)

		writeErr := func(err error) {
			writeResJSON(w, http.StatusInternalServerError, map[string]any{
				"error": err.Error(),
			})
		}

		knownErrors := []error{
			ErrOpeningDatabase,
			ErrDatabaseConnection,
			ErrPlaylistDirCreate,
			ErrTableCreation,
			ErrUpdatingFiles,
		}

		for _, knownErr := range knownErrors {
			if errors.Is(err, knownErr) {
				writeErr(knownErr)
				return
			}
		}

		writeErr(errors.New("Unknown Error"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Lists all tags
func ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := ListTagsAll()
	if err != nil {
		log.Printf("failed to list all tags %v", err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to query database",
		})

		return
	}

	writeResJSON(w, http.StatusOK, map[string]any{
		"tags": tags,
	})
}

// Lists tags associated to a file
func TagsByFileIDHandler(w http.ResponseWriter, r *http.Request) {
	param_file_id := r.PathValue("file_id")

	// Validating file id
	file_id, err := ValidateDbId(param_file_id)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Checking if file exists
	missing, err := CheckIDsInDB("file", []int64{file_id})
	if len(missing) != 0 {
		writeResJSON(w, http.StatusNotFound, map[string]any{
			"error": "file does not exist",
		})
		return
	}

	// Listing Tags
	tags, err := ListTagsByFileID(file_id)
	if err != nil {
		log.Printf("failed to list tags by file id %d: %v", file_id, err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to query database",
		})
		return
	}

	writeResJSON(w, http.StatusOK, map[string]any{
		"tags": tags,
	})
}

// Creates a tag
func CreateTagHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TagName string `json:"name"`
	}

	// Invalid JSON
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid JSON",
		})
		return
	}

	// name missing or empty
	if req.TagName == "" {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "name is required",
		})
		return
	}

	// Validating Name
	if err := ValidateTagName(req.TagName); err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Create the tag
	created_id, err := CreateTag(req.TagName)
	if err != nil {
		if errors.Is(err, ErrTagAlreadyExists) {
			writeResJSON(w, http.StatusConflict, map[string]any{
				"error": ErrTagAlreadyExists.Error(),
			})
			return
		}
		log.Printf("failed creating tag %s: %v", req.TagName, err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to insert into database",
		})
		return
	}

	writeResJSON(w, http.StatusCreated, map[string]any{
		"id": created_id,
	})
}

// Renames a tag
func RenameTagHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TagID   string `json:"tagID"`
		NewName string `json:"newName"`
	}

	// Invalid JSON
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid JSON",
		})
		return
	}

	// TagID is missing or empty
	if req.TagID == "" {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "tagID is required",
		})
		return
	}

	// newName is missing or empty
	if req.NewName == "" {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "newName is required",
		})
		return
	}

	// Validate TagID
	tagID, err := ValidateDbId(req.TagID)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "tagID is invalid",
		})
		return
	}

	// Validate newName
	if err := ValidateTagName(req.NewName); err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "newName is invalid",
		})
		return
	}

	// Check if tag exists
	missing, err := CheckIDsInDB("tag", []int64{tagID})
	if len(missing) != 0 {
		writeResJSON(w, http.StatusNotFound, map[string]any{
			"error": "tag not found",
		})
		return
	}

	// Rename the tag
	if err := RenameTag(tagID, req.NewName); err != nil {
		if errors.Is(err, ErrTagAlreadyExists) {
			writeResJSON(w, http.StatusConflict, map[string]any{
				"error": ErrTagAlreadyExists.Error(),
			})
			return
		}
		log.Printf("failed renaming tag with id %d to %s: %v", tagID, req.NewName, err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to update database",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Streaming a media file over http
func MediaHandler(w http.ResponseWriter, r *http.Request) {
	param_file_id := r.PathValue("file_id")

	// Validating file id
	file_id, err := ValidateDbId(param_file_id)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Getting the file from database
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
				"error": "Failed to query database",
			})
			return
		}
	}

	// Serve the file
	path := filepath.Join(dbPath, file.Name)
	http.ServeFile(w, r, path)
}
