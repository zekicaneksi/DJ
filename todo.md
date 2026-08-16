# TODO

### Notes

- Code the frontend and the backend completely separately. Write tests that will not need the other one to work for both.
    - In the end, test them together as well (integration test).
- Code the backend first.

### List

- Code the backend routes by design (design/routes_api.md), write their tests as well.
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
- Code the frontend
    - The design is more of a wireframe design. Make the real deal cooler.
