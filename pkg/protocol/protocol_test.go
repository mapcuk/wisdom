package protocol

import (
	"bytes"
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallengeRequest(t *testing.T) {
	buf := bytes.Buffer{}
	initMessageOnClient := NewMessage(ChallengeRequest, nil)
	require.NoError(t, initMessageOnClient.Write(&buf))
	initMessageOnServer, err := ReadMessage(&buf)
	require.NoError(t, err)
	assert.Equal(t, initMessageOnServer.GetBody(), initMessageOnClient.GetBody())
	assert.Equal(t, initMessageOnServer.Kind, initMessageOnClient.Kind)
}

func TestLookForSolution(t *testing.T) {
	easyNonce := Nonce{
		Zeros: 2,
		Salt:  []byte("abc"),
	}
	got, err := LookForSolution(easyNonce)
	require.NoError(t, err)
	assert.Equal(t, []byte("3\x9e\x00\x00"), got)
}

func TestWorkFlow(t *testing.T) {
	var testZerosLength uint = 2
	srv, clnt := net.Pipe()
	go func() {
		req := NewMessage(ChallengeRequest, nil)
		req.Write(clnt)
		resp, _ := ReadMessage(clnt)
		assert.Equal(t, ChallengeResponse, resp.Kind)
		nonce := Nonce{}
		err := json.Unmarshal(resp.GetBody(), &nonce)
		require.NoError(t, err)
		assert.Equal(t, testZerosLength, nonce.Zeros)
		assert.Len(t, nonce.Salt, 8)
		badSolution := NewMessage(ChallengeSolution, []byte("this is bad solution"))
		badSolution.Write(clnt)

		reportResp, _ := ReadMessage(clnt)
		assert.Equal(t, ErrorReport, reportResp.Kind)
		assert.Equal(t, "Wrong solution", string(reportResp.GetBody()))
		clnt.Close()
	}()

	HandleOnServer(srv, testZerosLength)
	srv.Close()
}
