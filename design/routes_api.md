# Routes

## Table of Contents

- [Frontend](#frontend)
  - [/choose-dir](#choose-dir)
  - [/](#-1)
- [Backend](#backend)
  - [Setup](#setup)
    - [POST /choose-dir](#post-choose-dir)
  - [Tags](#tags)
    - [GET /tags](#get-tags)
    - [GET /tags/{file_id}](#get-tagsfile_id)
    - [POST /create-tag](#post-create-tag)
    - [POST /rename-tag](#post-rename-tag)
    - [POST /delete-tag](#post-delete-tag)
    - [POST /update-tag](#post-update-tag)
  - [Files](#files)
    - [POST /search-files-by-tag](#post-search-files-by-tag)
    - [GET /media/{file_id}](#get-mediafile_id)
  - [Playlists](#playlists)
    - [POST /create-playlist](#post-create-playlist)

## Frontend

### /choose-dir

Choosing a directory for the app.

### /

Main page of the app.

## Backend

All routes will be prefixed with `/api`.

## Setup

### POST /choose-dir

Choose a directory and set up the database in the backend.

```text
{
    "dirPath": string
}
```

#### 204 - No Content

```text
Empty body
```

#### 400 - Bad Request

```json
{
  "error": "Invalid directory path"
}
```

#### 403 - Forbidden

```json
{
  "error": "Failed to create the directory"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to initialize database"
}
```

## Tags

### GET /tags

Get the list of all tags.

#### 200 - Success

```text
{
    "tags": []Tag { ID: int, Name: string }
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```

### GET /tags/{file_id}

Get the list of tags of a file by id.

#### 200 - Success

```text
{
    "tags": []Tag { ID: int, Name: string }
}
```

#### 400 - Bad Request

```json
{
  "error": "Invalid file id"
}
```

#### 404 - Not Found

```json
{
  "error": "File not found"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```

### POST /create-tag

Create a tag.

```text
{
    "name": string
}
```

#### 201 - Created

```text
{
    "id": int
}
```

#### 400 - Bad Request

```json
{
  "error": "Invalid tag"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```

### POST /rename-tag

Rename a tag.

```text
{
    "tagID": int,
    "newName": string
}
```

#### 204 - No Content

```text
Empty body
```

#### 400 - Bad Request

```json
{
  "error": "Invalid tagID/name"
}
```

#### 404 - Not Found

```json
{
  "error": "Tag not found"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```

### POST /delete-tag

Delete a tag.

```text
{
    "tagID": int
}
```

#### 204 - No Content

```text
Empty body
```

#### 400 - Bad Request

```json
{
  "error": "Invalid tag id"
}
```

#### 404 - Not Found

```json
{
  "error": "Tag not found"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```

### POST /update-tag

Updates tags of a file.

```text
{
    "fileID": int,
    "tagIDs": []int
}
```

#### 204 - No Content

```text
Empty body
```

#### 400 - Bad Request

```json
{
  "error": "Invalid tags ids/file id"
}
```

#### 404 - Not Found

```json
{
  "error": "Tag/File not found"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```

## Files

### POST /search-files-by-tag

Get the list of files from the backend by tag ids.

```text
{
    "tagIDs": []int
}
```

#### 200 - Success

```text
{
    "files": []File { ID: int, Name: string }
}
```

#### 400 - Bad Request

```json
{
  "error": "Invalid tag ids"
}
```

#### 404 - Not Found

```json
{
  "error": "x tags are missing"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```

### GET /media/{file_id}

Streams a file by id.

#### 200 - Success

```text
The file
```

#### 400 - Bad Request

```json
{
  "error": "Invalid file id"
}
```

#### 403 - Forbidden

```json
{
  "error": "Cannot stream the file"
}
```

#### 404 - Not Found

```json
{
  "error": "File not found"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```

## Playlists

### POST /create-playlist

Creates a playlist file.

```text
{
    "tagGroups": [][]TagGroup { TagIDs: []int, Amount: int }
}
```

#### 201 - Created

```text
{
    "name": string
}
```

#### 400 - Bad Request

```json
{
  "error": "Invalid tag groups"
}
```

#### 403 - Forbidden

```json
{
  "error": "Cannot create the file"
}
```

#### 404 - Not Found

```json
{
  "error": "x tags are missing"
}
```

#### 500 - Internal Server Error

```json
{
  "error": "Failed to query database"
}
```
