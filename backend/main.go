package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := setupServer()

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func setupServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /about", aboutHandler)

	return mux
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "About page")
}
