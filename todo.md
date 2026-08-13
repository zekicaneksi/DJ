# TODO

### Notes

- Code the frontend and the backend completely separately. Write tests that will not need the other one to work for both.
    - In the end, test them together as well (integration test).
- Code the backend first.

### List

- Change updating (Attaching/Detaching) tags
    - `POST /update-tag` route will send the tag ids a file should have.
        - Delete `AttachTag` and `DetachTag` functions from the backend. Instead have `UpdateTags` function which will have a database transaction that deletes all the tags of a file then adds the ones that came with the request.
- `GET /tags/{file_id}` route requires a function to get tags of a file by id 
- Better error messages from the backend functions and have more checks
- Unit tests that test every scenario
- Integration test that simulates normal usage, not every scenario
- Code the backend routes by design (design/routes_api.md)
- Code the frontend
