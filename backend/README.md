# DJ - Backend

### Technologies

- Go
- SQLite

### Environment Variables

```
DJ_BACKEND_PORT --- The backend port
```

## Development

### Test

```
go clean -testcache && go test -v -parallel 1 .
```

### Run

```
go run .
```

## Build

To create the executable

```
go build .
```