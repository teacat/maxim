package main

import (
	"log"

	"github.com/teacat/maxim"
)

func main() {
	log.Println("Running...")
	done := make(chan bool, 1)

	c, _, err := maxim.NewClient(&maxim.ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
	if err != nil {
		panic(err)
	}
	c.HandleMessage(func(_ *maxim.Client, msg string) {
		log.Println("received: " + msg)
		done <- true
	})
	err = c.Write("Hello")
	if err != nil {
		panic(err)
	}

	<-done
}
