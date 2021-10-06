package main

import (
	"github.com/kuberlogic/kuberlogic/modules/agent/client"
	"log"
	"os"
)

func main() {
	c, err := client.NewClient(
		os.Getenv("NAME"),
		os.Getenv("CONTROLLER_ADDR"),
		os.Getenv("INIT_ONLY") != "")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(c.Run())
}
