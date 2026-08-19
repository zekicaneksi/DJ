package main

import (
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {

	config, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Close the database before exiting the app
	closeDB := func() {
		if err := CloseDB(); err != nil {
			fmt.Println(err)
		}
	}
	defer closeDB()

	// Set up server
	handler := SetupServer()

	// Listen
	fmt.Println("Server running at " + config.Port)
	if err := http.ListenAndServe(":"+config.Port, handler); err != nil {
		fmt.Println(err)
	}

}
