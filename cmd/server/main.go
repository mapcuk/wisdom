package main

import (
	"context"
	slog "log"
	"math/rand"
	"time"

	"github.com/mapcuk/wisdom/internal/log"
	"github.com/mapcuk/wisdom/internal/server"

	"github.com/caarlos0/env/v9"
	"go.uber.org/zap"
)

type config struct {
	Addr    string `env:"ADDR" envDefault:"localhost:9000"`
	IsDebug bool   `env:"DEBUG"`
	Zeros   uint   `env:"ZEROS" envDefault:"3"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Fatalf("config parse %v", err)
	}
	rand.Seed(time.Now().UnixNano())
	logger := log.New(cfg.IsDebug)
	ctx := context.TODO() // NOTICE: migrate to WithCancel and implement graceful shutdown
	if err := server.Start(ctx, cfg.Addr, cfg.Zeros); err != nil {
		logger.Error("start server", zap.Error(err))
	}

}
