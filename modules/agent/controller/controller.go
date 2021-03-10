package controller

import (
	"log"
	"sync"
)

var Controller *AgentController

const globalControllerPort = 18888

var runOnce sync.Once

func Init() {
	runOnce.Do(func() {
		Controller = NewController(globalControllerPort)
		go log.Fatal(Controller.Run())
	})
}
