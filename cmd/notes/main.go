package main

import (
	"context"
	"flag"
	"log"
	"notes/internal/notes/app"
	"notes/internal/notes/controller/notes"
	"notes/internal/notes/storage/postgres"
	"notes/internal/pkg/config"
	"notes/internal/pkg/logger"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "./config/local.yaml", "config path")
}

func main() {
	flag.Parse()
	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalf("can not set up config error: %s", err.Error())
	}

	logg, err := logger.New(cfg.Env)
	if err != nil {
		log.Fatalf("can not set up logger error: %s", err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	ctxS, cancelS := context.WithTimeout(ctx, time.Second*5)
	defer cancelS()
	strPostgres, err := postgres.New(ctxS, cfg)
	if err != nil {
		logg.Error("storage initializing error", zap.Error(err))
	}

	a := app.NewApp(strPostgres)

	s, err := notes.New(notes.RestAPI, cfg, a, logg)
	if err != nil {
		logg.Fatal("service initializing failed", zap.String("error", err.Error()))
	}

	sG, err := notes.New(notes.GRPCAPI, cfg, a, logg)
	if err != nil {
		logg.Fatal("service initializing failed", zap.String("error", err.Error()))
	}

	go func() {
		<-ctx.Done()

		ctxS, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := s.Shutdown(ctxS); err != nil {
			logg.Error("can not shutdown REST server", zap.String("error", err.Error()))
		}

		if err := sG.Shutdown(ctxS); err != nil {
			logg.Error("can not shutdown GRPC server", zap.String("error", err.Error()))
		}

		logg.Info("server shutdown success")
	}()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		logg.Info("started REST API on", zap.String("address", cfg.Server.Host+cfg.Server.Port))
		if err := s.Start(ctx); err != nil {
			logg.Error("can not start server", zap.String("error", err.Error()))
		}
		wg.Done()
	}()

	go func() {
		logg.Info("started GRPC API on", zap.String("address", cfg.GRPCServer.Host+cfg.GRPCServer.Port))
		if err := sG.Start(ctx); err != nil {
			logg.Error("can not start GRPC server", zap.String("error", err.Error()))
		}
		wg.Done()
	}()
	wg.Wait()

	// autotls.RunWithContext()
}
