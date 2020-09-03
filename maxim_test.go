package maxim

import (
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	assert := assert.New(t)

	l, err := net.Listen("tcp", ":8080")
	assert.NoError(err)

	go func() {

		m := NewDefault()
		m.HandleMessage(func(s *Session, msg string) {
			assert.Equal("Hello", msg)

			err := s.Write(msg + ", world")
			assert.NoError(err)

		})
		http.HandleFunc("/ws", m.HandleRequest)

		err = http.Serve(l, nil)
		//assert.NoError(err)
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
