# TODO

## Notes

Code the frontend and the backend completely separately. Write tests that will not need the other one to work for both. In the end, test them together as well (integration test).

Code the backend first.

## List

### Repetitive Code in `TestWorkflow`

The helper principle applied in `TestWorkflowHandlers` can be used here as well.

### Listing Untagged Files

I forgot the route for listing untagged files. There is a function already written for it `ListFilesUntagged` and it's being tested too. But there is no route for it in the design, and its handler is not written.
- Add it to the design.
- Write the handler and the test for it.

Also in these tests, check for empty array;
- `TestFilesByTagHandler`
- `TestListFilesByTagIDs`

Fix the commented line in `TestWorkflowHandlers`

### Listing All Files

I forgot the route for listing all files. There is a function already written for it `ListFilesAll` and it's being tested too. But there is no route for it in the design and its handler is not written.
- Add it to the design.
- Write the handler and the test for it.

Fix the commented line in `TestWorkflowHandlers`

### Non-invasive DJ

Instead of creating the SQLite file and the Playlist directory and their files in the music directory, a safer and less permission required way would be to create these in the directory where the executable is.

Think about directory change as well. Let's say someone had already opened up a directory and did some tagging. If the user moves this directory then what?
- An idea is, anytime UpdateFiles function detects new/missing files, it can maybe ask to update the files or the directory path to the user?

I've thought about it, the only things I would need to change would be;
- the `InitDatabase` function to create the directories.
- the cleanup function would need some change for testing.
- the frontend's /choose-dir route would list the directories DJ has registered. 
- the operations to rename, change, add, etc. would have to be implemented both in the backend and frontend.

All of this requires a little work but worth in the end. It is important to keep the music directory as is. Do it. Check all the files afterwards just in case. Such a check was needed anyway. Don't forget to update the design.

Check if the tips in the `Multiple Connections` section of this document is still valid after the update.

### Frontend

Code the frontend. Note that the design is more of a wireframe design. Make the real thing cooler.

## New Features

After completing the project as it is designed, here are some things you can add.

### Multiple Connections

Currently, app works with only a single connection. Opening a new connection will close the previous one. If for whatever reason will you need multiple connections, here are some tips;

Create a type for database handling, such as
```Go
type DBConnection struct {
    DBPath string
    DBHandle *sql.DB
}
```

Create a map such as `var dbConnections map[string]DBConnection` the key will be an UUID, generated at the time of initialization.

Look for both `dbPath` and `dbHandle` in the source code. Wherever they are used instead need to use the map.

Except for the initialization handler, every handler will need to know which database the request came for. Therefor figure out a way to store the UUID of databases in the frontend. And access them in the handlers. If needed, change the `design/routes_api.md` accordingly.

### Deployment

Currently, the project is designed to be used locally. Though I don't think it would be difficult to deploy it to a server that has music directories which can be streamed remotely. Playlists could be created remotely too.
