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
	for {
		_, err := c.Read()
		if err != nil {
			panic(err)
		}
	}
}
