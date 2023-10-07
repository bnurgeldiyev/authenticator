package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"authenticator/config"
	"authenticator/internal/controller"
	"authenticator/internal/usecase"
	"authenticator/pkg/cache"
	"authenticator/pkg/postgres"
)

func Run(cfg *config.Config) {

	ctx := context.Background()
	c, err := cache.NewRedisService(ctx, cfg.Redis.URL, 1, 8, 256)
	if err != nil {
		log.Fatal().Err(err).Msg("app - Run - cache.NewRedisService")
	}

	pg, err := postgres.NewService(cfg)
	if err != nil {
		log.Panic().Err(err).Msg("app - Run - postgres.NewService")
		return
	}

	useCases := usecase.LoadUseCases(pg, c)

	userRouter := controller.NewUserRouter(useCases.UserUseCase)
	s := grpc.NewServer()
	controller.RegisterAuthServiceServer(s, userRouter)

	lis, err := net.Listen("tcp", ":"+cfg.Http.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("App - net.Listen")
		return
	}

	go setupSerer(s, lis)

	signalChan := make(chan os.Signal, 1)
	quitChan := make(chan interface{})
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	fmt.Println("<--START-SERVER-->", cfg.Http)
	for {
		select {
		case <-quitChan:
			log.Warn().Msg("quit channel closed, closing listener")
			s.Stop()

			err = lis.Close()
			if err != nil {
				log.Err(err).Msg("App - lis.Close()")
			}
			return
		case sig := <-signalChan:
			switch sig {
			case os.Interrupt, os.Kill, syscall.SIGTERM:
				log.Info().Msg("interrupt signal received, sending Quit signal")
				close(quitChan)
			default:
				log.Info().Msg("signal received")
			}
		}
	}
}

func setupSerer(srv *grpc.Server, lis net.Listener) {
	if err := srv.Serve(lis); err != nil {
		log.Fatal().Msg("App - setupServer - srv.Serve(list)")
		return
	}
}
