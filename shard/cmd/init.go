package cmd

import (
	"context"
	"prismarine/shard/server"
	"time"

	"github.com/charmbracelet/log"
)

func Execute() {
	log.SetLevel(log.DebugLevel)
	manager, err := server.NewManager(context.Background())
	if err != nil {
		log.Fatal("failed to initialize manager")
		return
	}

	for _, s := range manager.All() {
		log.Debugf("server %s loaded", s.Id())

		if err := s.Start(s.Context(), false, 0); err != nil {
			log.Error(err, "Failed to start server")
		}

	}

	log.Debug("Waiting...")
	time.Sleep(10 * time.Second)
	log.Debug("Executing again...")

	// for _, s := range manager.All() {
	// 	if err := s.HandlePowerAction(server.PowerActionStop, 0); err != nil {
	// 		log.Error(err, "Failed to stop server")
	// 	}
	// }
}
