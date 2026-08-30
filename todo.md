# TODO

### Notes

- Code the frontend and the backend completely separately. Write tests that will not need the other one to work for both.
    - In the end, test them together as well (integration test).
- Code the backend first.

### List

- A workflow test using only the API routes.
- Instead of creating the SQLite file and the Playlist directory and their files in the music directory, a safer and less permission required way would be to create these in the directory where the executable is. I've thought about it, the only things I would need to change would be the InitDatabase function to create the directories. Also the cleanup function would need some change for testing. And finally the frontend's /choose-dir route would list the directories DJ has registered. And the operations to rename, change, add, etc. would have to be implemented both in the backend and frontend. All of this requires a little work but worth in the end. It is important to keep the music directory as is. Do it. Check all the files afterwards just in case. Such a check was needed anyway. Don't forget to update the design.
    - Check if the tips in the `Multiple Connections` section of this document is still valid after the update.
- Code the frontend
    - The design is more of a wireframe design. Make the real deal cooler.

### Multiple Connections

Currently, app works with only a single connection. Opening a new connection will close the previous one. If for whatever reason will you need multiple connections, here are some tips;
- Create a type for database handling, such as
```Go
type DBConnection struct {
    DBPath string
    DBHandle *sql.DB
}
```
- Create a map such as `var dbConnections map[string]DBConnection` the key will be an UUID, generated at the time of initialization.
- Look for both `dbPath` and `dbHandle` in the source code. Wherever they are used instead need to use the map.
- Except for the initialization handler, every handler will need to know which database the request came for. Therefor figure out a way to store the UUID of databases in the frontend. And access them in the handlers. If needed, change the `design/routes_api.md` accordingly.