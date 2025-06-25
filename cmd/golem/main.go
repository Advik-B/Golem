package main

import (
	server2 "github.com/Advik-B/Golem/server"
	"go.uber.org/zap"
)

func main() {
	server2.InitLogger(true)
	defer server2.Log.Sync()

	server2.Log.Info("Starting Golem server on :25565...")
	srv := server2.NewServer("0.0.0.0:25565")
	if err := srv.Start(); err != nil {
		server2.Log.Fatal("Failed to start server", zap.Error(err))
	}
}
