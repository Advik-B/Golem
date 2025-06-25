package main

import (
	"fmt"
	"github.com/Advik-B/Golem/internal/server"
)

func main() {
	srv := server.NewServer("0.0.0.0:25565")
	fmt.Println("Starting Golem server on :25565...")
	err := srv.Start()
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
