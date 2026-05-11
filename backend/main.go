package main

import (
	"log"

	"social-network/backend/server"
)

func main() {
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
