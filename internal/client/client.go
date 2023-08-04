package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/mapcuk/wisdom/internal/log"
	"github.com/mapcuk/wisdom/pkg/protocol"

	"go.uber.org/zap"
)

func GetSomeWisdom(ctx context.Context, address string) (string, error) {
	logger := log.Get()
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	logger.Info("connect", zap.String("address", address))
	defer conn.Close()

	initRequest := protocol.NewMessage(protocol.ChallengeRequest, []byte(""))
	if err := initRequest.Write(conn); err != nil {
		return "", err
	}
	timeout := 15 * time.Second
	for {
		conn.SetDeadline(time.Now().Add(timeout))
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context is Done")
		default:
		}

		msg, err := protocol.ReadMessage(conn)
		if err != nil {
			return "", err
		}
		switch msg.Kind {
		case protocol.ChallengeResponse:
			nonce := protocol.Nonce{}
			err := json.Unmarshal(msg.GetBody(), &nonce)
			if err != nil {
				return "", err
			}
			solution, err := protocol.LookForSolution(nonce)
			if err != nil {
				return "", err
			}
			solutionResp := protocol.NewMessage(protocol.ChallengeSolution, solution)
			if err := solutionResp.Write(conn); err != nil {
				return "", err
			}
		case protocol.WordResponse:
			return string(msg.GetBody()), nil
		}
	}
}
