package main

import (
	"context"
	slog "log"
	"time"

	"github.com/caarlos0/env"
	"github.com/mapcuk/wisdom/internal/client"
	"github.com/mapcuk/wisdom/internal/log"

	"go.uber.org/zap"
)

type config struct {
	Addr string `env:"ADDR" envDefault:"localhost:9000"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Fatalf("config parse %v", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	logger := log.Get()
	wisdom, err := client.GetSomeWisdom(ctx, cfg.Addr)
	if err != nil {
		logger.Error("error during getting some wisdom", zap.Error(err))
	} else {
		logger.Info("got wisdom word", zap.String("wisdom", wisdom))
	}
}
