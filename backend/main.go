package main

import (
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {

	config, err := loadConfig()
	if err != nil {
		log.Println(err)
		return
	}

	// Close the database before exiting the app
	closeDB := func() {
		if err := CloseDB(); err != nil {
			log.Println(err)
		}
	}
	defer closeDB()

	// Set up server
	handler := SetupServer()

	// Listen
	log.Printf("Server running at %s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, handler); err != nil {
		log.Println(err)
	}

}
