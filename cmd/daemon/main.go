package main

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"authenticator/config"
	"authenticator/internal/app"
	"authenticator/pkg/logger"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	pwd, _ := os.Getwd()

	zLog, err := logger.New("info", filepath.Join(pwd, "logs"))
	if err != nil {
		zLog.Panic().Err(err).Msg("could not initialize Logger")
		return
	}

	zerolog.DefaultContextLogger = zLog
	log.Logger = *zLog

	app.Run(cfg)
}
