package server

import (
	"context"
	"fmt"
	"net"
	"github.com/mapcuk/wisdom/internal/log"
	"github.com/mapcuk/wisdom/pkg/protocol"

	"go.uber.org/zap"
)

func Start(ctx context.Context, addr string, zeros uint) error {
	logger := log.Get()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	logger.Info("Start server", zap.String("addr", addr))

	for {
		// TODO ctx.Done()
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept %v", err)
		}
		go protocol.HandleOnServer(conn, zeros)
	}
}
