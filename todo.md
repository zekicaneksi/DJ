# TODO

### Notes

- Code the frontend and the backend completely separately. Write tests that will not need the other one to work for both.
    - In the end, test them together as well (integration test).
- Code the backend first.

### List

- Unit tests that test every scenario
- Integration test that simulates normal usage, not every scenario
- Code the backend routes by design (design/routes_api.md)
    - Validate the input in these routes. Validation function is already written in `validation.go`;
        - CreateTag
        - RenameTag
    - Example for when checking validation error
    ```
    	if _, err := ValidateTagName(tagName); err != nil {
		    if err == ErrTagNameTooLong {"tag name is too long"}
		    else {"database error"}
	    }
        CreateTag(tagName)
    ```
- Code the frontend
