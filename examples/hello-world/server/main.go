package main

import (
	"log"
	"net"
	"net/http"

	"github.com/teacat/maxim"
)

func main() {
	m := maxim.NewDefault()
	m.HandleMessage(func(s *maxim.Session, msg string) {
		log.Println("reveived: " + msg)
		s.Write(msg + ", world")
	})

	log.Println("Running...")

	http.HandleFunc("/ws", m.HandleRequest)
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	err = http.Serve(l, nil)
	if err != nil {
		panic(err)
	}
}
