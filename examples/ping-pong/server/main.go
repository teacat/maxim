package main

import (
	"log"
	"net"
	"net/http"

	"time"

	"github.com/teacat/maxim"
)

func main() {
	m := maxim.NewDefault()
	m.HandleConnect(func(s *maxim.Session) {
		log.Printf("connected: %+v", s)
		for {
			<-time.After(time.Second * 1)
			log.Printf("ping: %+v", s)
			err := s.Ping()
			if err != nil {
				panic(err)
			}
		}
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
