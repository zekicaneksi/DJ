package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := SetupServer()

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func SetupServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", HomeHandler)
	mux.HandleFunc("GET /about", AboutHandler)

	return mux
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "About page")
}
