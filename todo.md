# TODO

## Notes

Code the frontend and the backend completely separately. Write tests that will not need the other one to work for both. In the end, test them together as well (integration test).

Code the backend first.

## List

### Non-invasive DJ

Instead of creating the SQLite file and the Playlist directory in the music directory, a safer and less permission required way would be to create these in the directory where the executable is.

#### Functionality To Be Added

In the <b>/choose-dir</b> page, there will be an "Add Directory" button which will open a popup and the user will choose a directory and a name and it will be added to the database.

The added directories will be listed in the <b>/choose-dir</b> page. The operations can be peformed on these will be;
- Remove
- Change path
- Rename
- Open
    - When a directory is chosen, the backend will check for missing files, if there are any, it will ask the user to either update the directory path, or update the database (remove missing files).

#### Plan of Action

Create the <b>/choose-dir</b> page in Figma as described in the <b>Functionality To Be Added</b> section.

Update all the relevant files in the `design` directory;
- Create a directory named <b>UI</b> and,
    - Export the newly created page into it.
    - Move `UI.png` and `UI.md` files into it.
    - Update `UI.md`.
- <b>Database</b> section in `README.md`.
    - The section should have 2 sub-sections;
        - Music Directory (current section)
        - Directory Listing
- Add the new routes to `routes_api.md`.
    - Also check all of the routes to see if there are
        - Overlapping routes that do the same thing.
        - Redundant routes.
        - Routes that need new parameters.

Update the source code. The updates made on the `routes_api.md` file should reflect the changes required in the backend.

Inspect all the files in the project just in case. Such a check was needed anyway.

Check if the tips in the `Multiple Connections` section of this document are still valid.

### Frontend

Code the frontend. Note that the UI design is a wireframe design. Make the real thing cooler.

## New Features

After completing the project as it is designed, here are some things you can add.

### Multiple Connections

Currently, app works with only a single connection (a single browser tab). Opening a new connection will close the previous one. And using the previous connection anyway might result in unexpected behavior. If for whatever reason will you need multiple connections, here are some tips;

Create a type for database handling, such as
```Go
type DBConnection struct {
    DBPath string
    DBHandle *sql.DB
}
```

Create a map such as `var dbConnections map[string]DBConnection` the key will be an UUID, generated at the time of initialization.

Look for both `dbPath` and `dbHandle` in the source code. Wherever they are used instead need to use the map.

Except for the initialization handler, every handler will need to know which database the request came for. Therefore figure out a way to store the UUID of databases in the frontend. And access them in the handlers. If needed, change the `design/routes_api.md` accordingly.

### Deployment

Currently, the project is designed to be used locally. Though I don't think it would be difficult to deploy it to a server that has music directories which can be streamed remotely. Playlists could be created remotely as well.
