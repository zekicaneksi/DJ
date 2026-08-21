package main

import (
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

func setUpLog() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	setUpLog()

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("err when loading config: %v", err)
	}

	// Close the database before exiting the app
	closeDB := func() {
		if err := CloseDB(); err != nil {
			log.Printf("error when closing db: %v", err)
		}
	}
	defer closeDB()

	// Set up server
	handler := SetupServer()

	// Listen
	log.Printf("Server running at %s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, handler); err != nil {
		log.Fatalf("error when listening: %v", err)
	}

}
