package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"img-sonurai-com/internal"
)

var Version = "dev"

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix // best for performance
	log.Logger = log.With().Str("version", Version).Logger()
	err := internal.Start()
	if err != nil {
		return
	}
}
