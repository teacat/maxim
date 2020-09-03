package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/teacat/maxim"
)

func main() {
	r := gin.Default()
	m := maxim.NewDefault()

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})
	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})
	m.HandleMessage(func(s *maxim.Session, msg string) {
		m.Write(msg)
	})
	r.Run(":8080")
}
