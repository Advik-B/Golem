package main

import (
	"log"

	"github.com/Advik-B/Golem/internal/server"
)

func main() {
	srv, err := server.New("0.0.0.0:25565")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	srv.Start()
}
