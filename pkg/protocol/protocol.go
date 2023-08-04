package protocol

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
	"github.com/mapcuk/wisdom/internal/log"

	"go.uber.org/zap"
)

var (
	ErrUnknown       = errors.New("Unknown request")
	ErrBadRequest    = errors.New("Bad request")
	ErrWrongSolution = errors.New("Wrong solution")
)

const (
	zeroCode = 48
	saltSize = 8
)

var words = []string{
	"You create your own opportunities.",
	"Never break your promises.",
	"You are never as stuck as you think you are.",
	"Happiness is a choice.",
	"Habits develop into character.",
	"Be happy with who you are.",
}

type MessageType uint8

const (
	ChallengeRequest MessageType = iota + 1
	ChallengeResponse
	ChallengeSolution
	WordResponse
	ErrorReport
)

/*
Protocol
Every request must be following
 - 3 magic bytes
 - 1 byte of meesage type
 - 2 bytes contains body size
 - rest of message is considered as body with certain size described above

 Workflow:
 1 Client sends ChallengeRequest: body is empty
 2 Server responds with ChallengeResponse: body = {nonce, zeros}
 3 Client calculates solution and sends ChallengeSolution: body = {solution}
 4 Server responds with WordResponse: body = {word}
*/

// magicHeader represents 3 bytes every requests must start with
var magicHeader = []byte{0xe1, 0xb7, 0x9c}

// Message info about request
type Message struct {
	Kind MessageType
	body []byte
}

func ReadMessage(r io.Reader) (*Message, error) {
	var hdr = make([]byte, 6)
	_, err := r.Read(hdr)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(hdr[:3], magicHeader) != 0 {
		return nil, ErrUnknown
	}
	msg := Message{}
	msg.Kind = MessageType(hdr[3])
	if msg.Kind > 5 {
		return nil, ErrBadRequest
	}
	bodyLength := binary.LittleEndian.Uint16(hdr[4:6])
	if bodyLength > 0 {
		msg.body = make([]byte, bodyLength)
		readBytes, err := r.Read(msg.body)
		if err != nil {
			return nil, fmt.Errorf("read error %v", err)
		}
		if readBytes < int(bodyLength) {
			return nil, ErrBadRequest
		}
	}
	return &msg, nil
}

func NewMessage(kind MessageType, body []byte) *Message {
	return &Message{Kind: kind, body: body}
}

func (m *Message) Write(w io.Writer) error {
	resp := bytes.Buffer{}
	resp.Write(magicHeader)
	resp.WriteByte(byte(m.Kind))
	bodyLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(bodyLength, uint16(len(m.body)))
	resp.Write(bodyLength)
	resp.Write(m.body)
	_, err := resp.WriteTo(w)
	return err
}

func (m *Message) GetBody() []byte {
	return m.body
}

func CheckSolution(solution []byte, n *Nonce) error {
	buf := bytes.NewBuffer(solution)
	buf.Write(n.Salt)
	result := Hash(buf.Bytes())
	foundZeros := 0
	for _, v := range result {
		if foundZeros == int(n.Zeros) {
			return nil
		}
		if v != zeroCode {
			return ErrWrongSolution
		}
		foundZeros += 1
	}
	return ErrWrongSolution
}

func LookForSolution(nonce Nonce) ([]byte, error) {
	var guess uint32 = 0
	buf := bytes.Buffer{}
	logger := log.Get()
	for {
		binary.Write(&buf, binary.LittleEndian, guess)
		if err := CheckSolution(buf.Bytes(), &nonce); err == nil {
			logger.Sugar().Infof("found solution %x", buf.Bytes())
			return buf.Bytes(), nil
		}
		buf.Reset()
		guess += 1
	}
}

type Nonce struct {
	Zeros     uint   `json:"zeros"`
	CreatedAt int64  `json:"createdAt"`
	Salt      []byte `json:"salt"`
}

func NewNonce(zeros uint) Nonce {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		log.Get().Error("salt randomize", zap.Error(err))
	}
	return Nonce{
		Zeros:     zeros,
		CreatedAt: time.Now().Unix(),
		Salt:      salt,
	}
}

func Hash(s []byte) []byte {
	h := sha256.New()
	h.Write(s)

	bs := h.Sum(nil)
	return bs
}

func HandleOnServer(conn net.Conn, zeros uint) {
	defer conn.Close()
	logger := log.Get()
	timeout := 15 * time.Second
	var nonce Nonce
	for {
		conn.SetDeadline(time.Now().Add(timeout))
		msg, err := ReadMessage(conn)
		if err != nil {
			logger.Error("read message", zap.Error(err))
			return
		}
		switch msg.Kind {
		case ChallengeRequest:
			nonce = NewNonce(zeros)
			body, err := json.Marshal(nonce)
			if err != nil {
				logger.Error("nonce marshall", zap.Error(err))
				return
			}
			challResp := NewMessage(ChallengeResponse, body)
			if err := challResp.Write(conn); err != nil {
				logger.Error("message Write", zap.Error(err))
				return
			}
		case ChallengeSolution:
			if err := CheckSolution(msg.GetBody(), &nonce); err != nil {
				errResp := NewMessage(ErrorReport, []byte(err.Error()))
				errResp.Write(conn)
				return
			} else {
				wisdomWord := words[rand.Intn(len(words))]
				wordResp := NewMessage(WordResponse, []byte(wisdomWord))
				wordResp.Write(conn)
				return
			}
		}
	}
}
