package main

import (
	"log"

	"github.com/teacat/maxim"
)

func main() {
	log.Println("Running...")

	c, _, err := maxim.NewClient(&maxim.ClientConfig{
		Address: "ws://localhost:8080/ws",
	})
	if err != nil {
		panic(err)
	}
	err = c.Write("Hello")
	if err != nil {
		panic(err)
	}
	msg, err := c.Read()
	if err != nil {
		panic(err)
	}
	log.Println(msg)
}
