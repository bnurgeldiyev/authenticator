package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"authenticator/auth"
	"authenticator/config"
	"authenticator/internal/usecase"
	"authenticator/pkg/httpserver"
	"authenticator/pkg/postgres"
)

type srv struct {
	auth.AuthServiceServer
}

func Run(cfg *config.Config) {

	//ctx := context.Background()
	//c, err := cache.NewRedisService(ctx, cfg.Redis.URL, 1, 8, 256)
	//if err != nil {
	//	log.Panic("app - Run - cache.NewRedisService: %w", err)
	//}

	pg, err := postgres.NewService(cfg)
	if err != nil {
		log.Panic().Err(err).Msg("app - Run - postgres.NewService")
		return
	}

	useCases := usecase.LoadUseCases(pg)

	fmt.Println(useCases)

	s := grpc.NewServer()
	auth.RegisterAuthServiceServer(s, &srv{})

	handler := mux.NewRouter()
	httpServer := httpserver.New(handler, httpserver.Port(cfg.Http.Host, cfg.Http.Port))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	fmt.Println("<--START-SERVER-->", cfg.Http)
	select {
	case s := <-interrupt:
		err = httpServer.Shutdown()
		if err != nil {
			log.Err(err).Msg("app - Run - httpServer.Shutdown")
			return
		}

		log.Info().Msg("app - Run - signal: " + s.String())
		return
	case err = <-httpServer.Notify():
		log.Err(err).Msg("app - Run - httpServer.Notify")
		return
	}
}
