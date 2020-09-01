package main

import (
	"log"
	"net"
	"net/http"

	"github.com/teacat/maxim"
)

func main() {
	m := maxim.NewDefault()
	m.HandleConnect(func(s *maxim.Session) {
		log.Printf("connected: %+v", s)
	})
	m.HandleClose(func(s *maxim.Session, status maxim.CloseStatus, msg string) error {
		log.Printf("closed: %+v, %+v, %+v", s, status, msg)
		return nil
	})
	m.HandleDisconnect(func(s *maxim.Session) {
		log.Printf("disconnected: %+v", s)
	})
	m.HandleError(func(s *maxim.Session, err error) {
		log.Printf("error: %+v, %+v", s, err)
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
