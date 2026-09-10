package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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
	mux.HandleFunc("GET /all-files", AllFilesHandler)
	mux.HandleFunc("POST /search-files-by-tag", FilesByTagHandler)
	mux.HandleFunc("GET /media/{file_id}", MediaHandler)
	mux.HandleFunc("POST /create-playlist", CreatePlaylistHandler)

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

// Decodes JSON into a given generic struct
func decodeJSON[T any](w http.ResponseWriter, r *http.Request, dst *T) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid request body",
		})
		return false
	}

	return true
}

// Wrapper around CheckIDsInDB to reduce repetitive code
func IDsExist(
	w http.ResponseWriter,
	table string,
	typeOfID string, // Type of missing id to send to user. Do not use "table" parameter to prevent table name leakage to user.
	ids []int64,
) bool {
	missing, err := CheckIDsInDB(table, ids)
	if err != nil {
		log.Printf("error when checking %s IDs %v: %v", table, ids, err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "error when querying database",
		})
		return false
	}

	if len(missing) != 0 {
		writeResJSON(w, http.StatusNotFound, map[string]any{
			"error": fmt.Sprintf("%s not found: %v", typeOfID, missing),
		})
		return false
	}

	return true
}

// Takes the http request's struct variable, and checks if every field is provided or not
func validateRequiredFields(w http.ResponseWriter, v any) bool {
	rv := reflect.ValueOf(v)
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		structField := rt.Field(i)

		// Remove options such as ",omitempty".
		jsonName := structField.Tag.Get("json")
		if comma := strings.IndexByte(jsonName, ','); comma >= 0 {
			jsonName = jsonName[:comma]
		}

		// nil pointer means the field wasn't provided.
		if field.Kind() == reflect.Pointer {
			if field.IsNil() {
				writeResJSON(w, http.StatusBadRequest, map[string]any{
					"error": fmt.Sprintf("%s is required", jsonName),
				})
				return false
			}

			field = field.Elem()
		}

		// Empty strings are considered unprovided.
		if field.Kind() == reflect.String && field.Len() == 0 {
			writeResJSON(w, http.StatusBadRequest, map[string]any{
				"error": fmt.Sprintf("%s is required", jsonName),
			})
			return false
		}
	}

	return true
}

// Checks if ID is int64 with PathValues
func checkPathValueInt64(w http.ResponseWriter, pathValue string) (int64, error) {
	file_id, err := strconv.ParseInt(pathValue, 10, 64)
	if err != nil {
		writeResJSON(w, http.StatusBadRequest, map[string]any{
			"error": "invalid int64 pathValue",
		})
		return 0, err
	}
	return file_id, nil
}

// Choose a directory and set up the database in the backend.
func ChooseDirHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DirPath *string `json:"dirPath"`
	}

	// Invalid Request Body
	if !decodeJSON(w, r, &req) {
		return
	}

	// Check fields
	if !validateRequiredFields(w, req) {
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
	file_id, err := checkPathValueInt64(w, param_file_id)
	if err != nil {
		return
	}

	// Checking if file exists
	if !IDsExist(w, "file", "file", []int64{file_id}) {
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
	if !decodeJSON(w, r, &req) {
		return
	}

	// Check fields
	if !validateRequiredFields(w, req) {
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
	if !decodeJSON(w, r, &req) {
		return
	}

	// Check fields
	if !validateRequiredFields(w, req) {
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
	if !IDsExist(w, "tag", "tag", []int64{*req.TagID}) {
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
	if !decodeJSON(w, r, &req) {
		return
	}

	// Check fields
	if !validateRequiredFields(w, req) {
		return
	}

	// Check if tag exists
	if !IDsExist(w, "tag", "tag", []int64{*req.TagID}) {
		return
	}

	// Delete the tag
	if err := DeleteTag(*req.TagID); err != nil {
		log.Printf("error when deleting tag %d: %v", *req.TagID, err)
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
	if !decodeJSON(w, r, &req) {
		return
	}

	// Check fields
	if !validateRequiredFields(w, req) {
		return
	}

	// File not found
	if !IDsExist(w, "file", "file", []int64{*req.FileID}) {
		return
	}

	// Tag not found
	if !IDsExist(w, "tag", "tag", *req.TagIDs) {
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

// Returns all files
func AllFilesHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ListFilesAll()
	if err != nil {
		log.Printf("failed to list all files %v", err)

		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to query database",
		})
		return
	}

	writeResJSON(w, http.StatusOK, map[string]any{
		"files": files,
	})
}

// Returns files that have the given tags
func FilesByTagHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TagIDs *[]int64 `json:"tagIDs"`
	}

	// Invalid Request Body
	if !decodeJSON(w, r, &req) {
		return
	}

	// Check fields
	if !validateRequiredFields(w, req) {
		return
	}

	// Tag not found
	if !IDsExist(w, "tag", "tag", *req.TagIDs) {
		return
	}

	// Get files
	files, err := ListFilesByTagIDs(*req.TagIDs)
	if err != nil {
		if errors.Is(err, ErrDuplicateID) {
			writeResJSON(w, http.StatusBadRequest, map[string]any{
				"error": ErrDuplicateID.Error(),
			})
			return
		}
		log.Printf("error when listing files with tags %v: %v", *req.TagIDs, err)
		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "error when querying database",
		})
		return
	}

	writeResJSON(w, http.StatusOK, map[string]any{
		"files": files,
	})
}

// Streaming a media file over http
func MediaHandler(w http.ResponseWriter, r *http.Request) {
	param_file_id := r.PathValue("file_id")

	// Validating file id
	file_id, err := checkPathValueInt64(w, param_file_id)
	if err != nil {
		return
	}

	// Getting the file from database
	var file File

	err = dbHandle.QueryRow(
		"SELECT id, name FROM file WHERE id = ?",
		file_id,
	).Scan(&file.ID, &file.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeResJSON(w, http.StatusNotFound, map[string]any{
				"error": "File not found",
			})
			return
		}
		log.Printf("error when querying database to stream file with id %d: %v", file_id, err)
		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Failed to query database",
		})
		return
	}

	// Serve the file
	path := filepath.Join(dbPath, file.Name)
	http.ServeFile(w, r, path)
}

// Creates a playlist
func CreatePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TagGroups *[]TagGroup `json:"tagGroups"`
	}

	// Invalid Request Body
	if !decodeJSON(w, r, &req) {
		return
	}

	// Check fields
	if !validateRequiredFields(w, req) {
		return
	}

	// Checking if all tag IDs exist
	var idsToCheck []int64
	for _, tagGroup := range *req.TagGroups {
		idsToCheck = append(idsToCheck, tagGroup.TagIDs...)
	}

	if !IDsExist(w, "tag", "tag", idsToCheck) {
		return
	}

	// Create the playlist
	playlistFilePath, err := CreatePlaylist(*req.TagGroups)
	if err != nil {
		// Validation Errors
		writeErr := func(err error) {
			writeResJSON(w, http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
		}

		knownErrors := []error{
			ErrTagGroupEmptyArr,
			ErrTagGroupEmpty,
			ErrTagGroupAmount,
			ErrTagGroupDuplicateID,
		}

		for _, knownErr := range knownErrors {
			if errors.Is(err, knownErr) {
				writeErr(knownErr)
				return
			}
		}

		// File operation errors
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			log.Printf("error when creating a playlist for %s, with %v: %v", dbPath, *req.TagGroups, pathErr)
			writeResJSON(w, http.StatusForbidden, map[string]any{
				"error": "Cannot create the file",
			})
			return
		}

		log.Printf("error when creating a playlist for %s, with %v: %v", dbPath, *req.TagGroups, err)
		writeResJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "Unknown Error",
		})
		return
	}

	writeResJSON(w, http.StatusCreated, map[string]any{
		"name": playlistFilePath,
	})
}
