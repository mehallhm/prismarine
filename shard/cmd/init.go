package cmd

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"prismarine/shard/service"
)

func Execute() {
	initLogging()
	log.Debug().Msg("Running in Debug mode")

}

func initLogging() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	manager, err := service.NewManager()
	if err != nil {
		log.Fatal().Msg("Failed to initialize manager")
	}

}
