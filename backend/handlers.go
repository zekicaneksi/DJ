package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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
	mux.HandleFunc("POST /delete-tag", DeleteTagHandler)
	mux.HandleFunc("POST /update-tag", UpdateTagHandler)
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
		DirPath *string `json:"dirPath"`
	}

	// Invalid Request Body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
		return
	}

	// dirPath missing or empty
	if req.DirPath == nil || len(*req.DirPath) == 0 {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "dirPath is required",
		})
		return
	}

	// Initialize database
	if err := InitDatabase(*req.DirPath); err != nil {
		log.Printf("failed to initialize db for path %s: %v", *req.DirPath, err)

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
	file_id, err := strconv.ParseInt(param_file_id, 10, 64)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "file_id is not valid int64",
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
		TagName *string `json:"name"`
	}

	// Invalid Request Body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
		return
	}

	// name missing or empty
	if req.TagName == nil || len(*req.TagName) == 0 {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "name is required",
		})
		return
	}

	// Validating Name
	if err := ValidateTagName(*req.TagName); err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Create the tag
	created_id, err := CreateTag(*req.TagName)
	if err != nil {
		if errors.Is(err, ErrTagAlreadyExists) {
			writeResJSON(w, http.StatusConflict, map[string]any{
				"error": ErrTagAlreadyExists.Error(),
			})
			return
		}
		log.Printf("failed creating tag %s: %v", *req.TagName, err)

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
		TagID   *int64  `json:"tagID"`
		NewName *string `json:"newName"`
	}

	// Invalid Request Body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
		return
	}

	// TagID is missing
	if req.TagID == nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "tagID is required",
		})
		return
	}

	// newName is missing or empty
	if req.NewName == nil || len(*req.NewName) == 0 {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "newName is required",
		})
		return
	}

	// Validate newName
	if err := ValidateTagName(*req.NewName); err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	// Check if tag exists
	missing, err := CheckIDsInDB("tag", []int64{*req.TagID})
	if len(missing) != 0 {
		writeResJSON(w, http.StatusNotFound, map[string]any{
			"error": "tag not found",
		})
		return
	}

	// Rename the tag
	if err := RenameTag(*req.TagID, *req.NewName); err != nil {
		if errors.Is(err, ErrTagAlreadyExists) {
			writeResJSON(w, http.StatusConflict, map[string]any{
				"error": ErrTagAlreadyExists.Error(),
			})
			return
		}
		log.Printf("failed renaming tag with id %d to %s: %v", *req.TagID, *req.NewName, err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to update database",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Deletes a tag
func DeleteTagHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TagID *int64 `json:"tagID"`
	}

	// Invalid Request Body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
		return
	}

	// TagID is missing
	if req.TagID == nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "tagID is required",
		})
		return
	}

	// Check if tag exists
	missing, err := CheckIDsInDB("tag", []int64{*req.TagID})
	if len(missing) != 0 {
		writeResJSON(w, http.StatusNotFound, map[string]any{
			"error": "tag not found",
		})
		return
	}

	// Delete the tag
	if err := DeleteTag(*req.TagID); err != nil {
		log.Printf("error when deleting tag %d: %v", req.TagID, err)
		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "internal error when deleting tag",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Updates a file's tags
func UpdateTagHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FileID *int64   `json:"fileID"`
		TagIDs *[]int64 `json:"tagIDs"`
	}

	// Invalid Request Body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
		return
	}

	// TagIDs missing
	if req.TagIDs == nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "tagIDs is required",
		})
		return
	}

	// fileID missing
	if req.FileID == nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "fileID is required",
		})
		return
	}

	// File not found
	missing, err := CheckIDsInDB("file", []int64{*req.FileID})
	if len(missing) != 0 {
		writeResJSON(w, http.StatusNotFound, map[string]any{
			"error": "file not found",
		})
		return
	}

	// Tag not found
	missing, err = CheckIDsInDB("tag", *req.TagIDs)
	if len(missing) != 0 {
		writeResJSON(w, http.StatusNotFound, map[string]any{
			"error": fmt.Sprintf("tag not found: %v", missing),
		})
		return
	}

	// Update
	if err := UpdateTags(*req.FileID, *req.TagIDs); err != nil {
		log.Printf("error when updating tag for %d with %v: %v", *req.FileID, *req.TagIDs, err)
		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "error when updating tags",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Streaming a media file over http
func MediaHandler(w http.ResponseWriter, r *http.Request) {
	param_file_id := r.PathValue("file_id")

	// Validating file id
	file_id, err := strconv.ParseInt(param_file_id, 10, 64)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid file_id",
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
