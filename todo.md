# TODO

### Notes

- Code the frontend and the backend completely separately. Write tests that will not need the other one to work for both.
    - In the end, test them together as well (integration test).
- Code the backend first.

### List

- Integration test that simulates intended usage, not every scenario. After every operation, check if it was done correctly or not.
    - Use the `workflow_test.go` file 
- Code the backend routes by design (design/routes_api.md)
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
