# Design

<b>This is only the initial design and once the coding starts, the design here will not be updated.</b>

## Requirements

- Tag the music files to categorize them.
    - A file can have multiple tags.

- Filter music files by one or multiple tags.
- Create random playlists by tags.

## Technologies

- Database -> SQLite
- Frontend -> mainly React
- Backend -> Go

## Database

```sql
CREATE TABLE IF NOT EXISTS file (
    id INTEGER PRIMARY KEY NOT NULL,
    name varchar(255) UNIQUE NOT NULL
);
```
```sql
CREATE TABLE IF NOT EXISTS tag (
    id INTEGER PRIMARY KEY NOT NULL,
    name varchar(255) UNIQUE NOT NULL
);
```
```sql
CREATE TABLE IF NOT EXISTS file_tag (
    file_id INT NOT NULL,
    tag_id INT NOT NULL,
    FOREIGN KEY(file_id) REFERENCES file(id) ON DELETE CASCADE,
    FOREIGN KEY(tag_id) REFERENCES tag(id) ON DELETE CASCADE,
    UNIQUE(file_id, tag_id)
);
```
