package main

import (
	"agent/command"
	"agent/controller"
	"log"
)

func main() {
	ctrl := controller.NewController(18888)

	testCommand, err := command.NewCommand("echo", "echo", "forever", "whatever")
	if err != nil {
		log.Fatal(err)
	}
	ctrl.InitCommands.Add("test", testCommand)

	testCommand2, err := command.NewCommand("demo", "./demo.sh")
	if err != nil {
		log.Fatal(err)
	}
	ctrl.Commands.Add("test", testCommand2)

	log.Fatal(ctrl.Run())
}