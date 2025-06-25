package main

import (
    "github.com/Advik-B/Golem/internal/server"
    "go.uber.org/zap"
)

func main() {
    server.InitLogger(true)
    defer server.Log.Sync()

    server.Log.Info("Starting Golem server on :25565...")
    srv := server.NewServer("0.0.0.0:25565")
    if err := srv.Start(); err != nil {
        server.Log.Fatal("Failed to start server", zap.Error(err))
    }
}