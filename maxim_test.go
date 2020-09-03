package maxim

import (
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	assert := assert.New(t)

	done := make(chan string)

	m := NewDefault()
	m.HandleMessage(func(s *Session, msg string) {
		done <- msg
	})
	http.HandleFunc("/ws", m.HandleRequest)

	go func() {
		l, err := net.Listen("tcp", ":8080")
		assert.NoError(err)

		err = http.Serve(l, nil)
		assert.NoError(err)
	}()

	c, _, err := NewClient(&ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
	assert.NoError(err)

	err = c.Write("foo")
	assert.NoError(err)

	assert.Equal("foo", <-done)
}
