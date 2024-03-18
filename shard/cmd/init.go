package cmd

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"prismarine/shard/server"
)

func Execute() {
	initLogging()
	log.Debug().Msg("Running in Debug mode")

}

func initLogging() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	manager, err := server.NewManager(context.Background())
	if err != nil {
		log.Fatal().Msg("Failed to initialize manager")
	}

	for _, s := range manager.All() {
		log.Debug().Msgf("server %s loaded", s.Id())

		if err := s.CreateInstance(); err != nil {
			log.Fatal().Err(err).Msg("We failed")
		}

	}

}
