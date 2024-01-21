package main

import (
	"log"

	"v/internal/server"
)

func main() {
	// Run the server and handle any potential errors
	if err := server.Run(); err != nil {
		// Log a fatal error and exit if the server fails to start
		log.Fatalln(err.Error())
	}
}
