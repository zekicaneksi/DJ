# TODO

### Notes

- Code the frontend and the backend completely separately. Write tests that will not need the other one to work for both.
    - In the end, test them together as well (integration test).
- Code the backend first.

### List

- Code the backend routes by design (design/routes_api.md), write their tests as well. (One route is already written "ChooseDirHandler", you can take it as an example.)
    - Validate the input in these routes. Validation functions are already written in `validation.go`;
        - CreateTag
        - RenameTag
        - CreatePlaylist
    - Example for when checking validation error
    ```
    	if _, err := ValidateTagName(tagName); err != nil {
		    if err == ErrTagNameTooLong {"tag name is too long"}
		    else {"database error"}
	    }
        CreateTag(tagName)
    ```
- Instead of creating the SQLite file and the Playlist directory and their files in the music directory, a safer and less permission required way would be to create these in the directory where the executable is. I've thought about it, the only things I would need to change would be the InitDatabase function to create the directories. Also the cleanup function would need some change for testing. And finally the frontend's /choose-dir route would list the directories DJ has registered. And the operations to rename, change, add, etc. would have to be implemented both in the backend and frontend. All of this requires a little work but worth in the end. It is important to keep the music directory as is. Do it. Check all the files afterwards just in case. Such a check was needed anyway.
- Code the frontend
    - The design is more of a wireframe design. Make the real deal cooler.
