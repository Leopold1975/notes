package main

import (
	"context"
	"flag"
	"log"
	notesgrpcclient "notes/internal/pkg/clients/notesGRPCclient"
	"notes/internal/pkg/config"
	"notes/internal/pkg/logger"
	"notes/internal/publisher"
	"notes/internal/publisher/kafkapublisher"
	"os/signal"
	"syscall"
	"time"
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

	kp, err := kafkapublisher.New(cfg.Kafka)
	if err != nil {
		logg.Fatalf("can not set up kafka publisher error %w", err)
	}

	gcf, err := notesgrpcclient.New(cfg.GRPCServer)
	if err != nil {
		logg.Fatalf("can not set up grpc fetcher error %w", err)
	}

	kb, err := publisher.New(kp, gcf, logg)
	if err != nil {
		logg.Fatal("can not set up kafka publisher", err)
	}
	logg.Info("Started publisher")

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	err = kb.Publish(ctx)
	if err != nil {
		logg.Error("can not publish messages", err)
	}

	for {
		select {
		case <-ctx.Done():
			logg.Info("Shutdown publisher")
			cancel()
			ctxS, cancelS := context.WithTimeout(context.Background(), time.Second*5)
			defer cancelS()
			kb.Shutdown(ctxS)
			break
		case <-time.After(time.Minute * 3):
			go func() {
				err = kb.Publish(ctx)
				if err != nil {
					logg.Error("can not publish messages", err)
				}
			}()
		}
	}
}
