package main

import (
	"github.com/kuberlogic/kuberlogic/modules/agent/command"
	"github.com/kuberlogic/kuberlogic/modules/agent/controller"
	"log"
)

func main() {
	controller.Init(18888)

	testCommand, err := command.NewCommand("echo", "echo", "forever", "whatever")
	if err != nil {
		log.Fatal(err)
	}
	controller.Controller.InitCommands.Add("test", testCommand)

	testCommand2, err := command.NewCommand("demo", "./demo.sh")
	if err != nil {
		log.Fatal(err)
	}
	controller.Controller.Commands.Add("test", testCommand2)
}
