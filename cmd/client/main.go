package main

import (
	"context"

	"time"
	"github.com/mapcuk/wisdom/internal/client"
	"github.com/mapcuk/wisdom/internal/log"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	logger := log.Get()
	wisdom, err := client.GetSomeWisdom(ctx, "localhost:9000")
	if err != nil {
		logger.Error("error during getting some wisdom", zap.Error(err))
	} else {
		logger.Info("got wisdom word", zap.String("wisdom", wisdom))
	}
}
