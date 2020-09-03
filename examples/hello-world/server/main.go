package main

import (
	"net/http"

	"github.com/teacat/maxim"
)

func main() {
	m := maxim.NewDefault()
	m.HandleMessage(func(s *maxim.Session, msg string) {
		s.Write(msg + ", world")
	})
	http.HandleFunc("/ws", m.HandleRequest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
