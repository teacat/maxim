package maxim

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	assert := assert.New(t)

	l, err := net.Listen("tcp", ":8080")
	assert.NoError(err)

	m := NewDefault()
	m.HandleMessage(func(s *Session, msg string) {
		assert.Equal("Hello", msg)
		err := s.Write(msg + ", world")
		assert.NoError(err)
	})

	go func() {
		http.HandleFunc("/ws", m.HandleRequest)
		http.Serve(l, nil)
	}()

	c, _, err := NewClient(&ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
	assert.NoError(err)

	err = c.Write("Hello")
	assert.NoError(err)

	msg, err := c.Read()
	assert.NoError(err)
	assert.Equal("Hello, world", msg)

	err = c.Close()
	assert.NoError(err)

	err = l.Close()
	assert.NoError(err)

}

type MyHandler struct {
	a *assert.Assertions

	hasMessage       bool
	hasMessageBinary bool
	hasError         bool
	hasClose         bool
	hasDisconnect    bool
	hasConnect       bool
}

func (h *MyHandler) HandleMessage(s *Session, msg string) {
	h.hasMessage = true
	h.a.Equal("Hello", msg)

	err := s.Write(msg + ", world")
	h.a.NoError(err)
}
func (h *MyHandler) HandleMessageBinary(s *Session, msg []byte) {
	h.hasMessageBinary = true
	h.a.Equal("Hello", string(msg))

	err := s.WriteBinary([]byte(string(msg) + ", world"))
	h.a.NoError(err)
}
func (h *MyHandler) HandleError(s *Session, err error) {
	h.hasError = true
}
func (h *MyHandler) HandleClose(s *Session, c CloseStatus, msg string) error {
	h.hasClose = true
	return nil
}
func (h *MyHandler) HandleDisconnect(s *Session) {
	h.hasDisconnect = true
}
func (h *MyHandler) HandleConnect(s *Session) {
	h.hasConnect = true
}

func TestHandle(t *testing.T) {
	assert := assert.New(t)
	h := &MyHandler{a: assert}

	l, err := net.Listen("tcp", ":8080")
	assert.NoError(err)

	m := NewDefault()
	m.Handle(h)

	go func() {
		http.HandleFunc("/ws", m.HandleRequest)
		http.Serve(l, nil)
	}()

	c, _, err := NewClient(&ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
	assert.NoError(err)

	err = c.Write("Hello")
	assert.NoError(err)

	msg, err := c.Read()
	assert.NoError(err)
	assert.Equal("Hello, world", msg)

	err = c.WriteBinary([]byte("Hello"))
	assert.NoError(err)

	msgBin, err := c.ReadBinary()
	assert.NoError(err)
	assert.Equal([]byte("Hello, world"), msgBin)

	err = c.Close()
	assert.NoError(err)

	// 等待 Close 與 Disconnect。
	<-time.After(1 * time.Second)

	assert.True(h.hasMessage)
	assert.True(h.hasMessageBinary)
	//assert.True(h.hasError)
	assert.True(h.hasClose)
	assert.True(h.hasDisconnect)
	assert.True(h.hasConnect)

	err = l.Close()
	assert.NoError(err)
}
