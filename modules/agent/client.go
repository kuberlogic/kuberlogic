package main

import (
	"agent/client"
	"log"
	"os"
)

func main() {
	c, err := client.NewClient(
		os.Getenv("NAME"),
		"localhost:18888",
		os.Getenv("INIT_ONLY") != "")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(c.Run())
}