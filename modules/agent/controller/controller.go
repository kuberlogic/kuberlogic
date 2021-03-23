package controller

import (
	"sync"
)

var Controller *AgentController

var runOnce sync.Once

func Init(port int) {
	runOnce.Do(func() {
		Controller = NewController(port)
		go Controller.Run()
	})
}
