package controller

import (
	"sync"
)

var Controller *AgentController

const globalControllerPort = 18888

var runOnce sync.Once

func Init() {
	runOnce.Do(func() {
		Controller = NewController(globalControllerPort)
		go Controller.Run()
	})
}
