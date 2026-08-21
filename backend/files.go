package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Gets the list of media files from a directory
func GetDirMediaFiles(path string) ([]string, error) {

	var fileNames []string

	// Allowed media file extensions
	fileExtensions := []string{
		"mp4", "mkv", "avi", "mov", "wmv",
		"flv", "webm", "mp3", "aac", "flac", "wav",
	}

	// Get files
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Making a map for faster look-up
	fileExtensionsMap := make(map[string]struct{}, len(fileExtensions))
	for _, ext := range fileExtensions {
		fileExtensionsMap[ext] = struct{}{}
	}

	// Filter the files from directories and non-media files
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

	return fileNames, nil
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
