package main

import (
	"github.com/Advik-B/Golem/log"
	"github.com/Advik-B/Golem/server"

	// Blank imports to trigger packet registration via init()
	_ "github.com/Advik-B/Golem/protocol/common"
	_ "github.com/Advik-B/Golem/protocol/configuration"
	_ "github.com/Advik-B/Golem/protocol/handshake"
	_ "github.com/Advik-B/Golem/protocol/login"
	_ "github.com/Advik-B/Golem/protocol/status"
	// _ "github.com/Advik-B/Golem/protocol/game" will be added later
)

func main() {
	log.ReplaceGlobals()

	srv := server.NewServer("tcp://:25565")
	if err := srv.Run(); err != nil {
		log.Logger.Fatal("Server failed to run", log.Error(err))
	}
}
