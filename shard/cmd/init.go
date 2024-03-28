package cmd

import (
	"context"
	"prismarine/shard/manager"
	"prismarine/shard/router"

	"github.com/charmbracelet/log"
)

func Execute() {
	log.SetLevel(log.DebugLevel)
	manager, err := manager.NewManager(context.Background())
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

	routes := router.Create(manager)
	routes.Listen(":3000")

	// log.Debug("Waiting...")
	// time.Sleep(10 * time.Second)
	// log.Debug("Executing again...")

	// for _, s := range manager.All() {
	// 	if err := s.WaitForStop(s.Context(), 30, true, false, 0); err != nil {
	// 		log.Error(err, "Failed to stop server")
	// 	}
	// }
}
