package klevent

import (
	"fmt"
	"sync"
)

const (
	eventNameField  = "kuberlogic.com/event-name"
	eventValueField = "kuberlogic.com/event-value"
)

type Controller struct {
	mu sync.Mutex

	ProcessedQ *EventQ
	HandlersQ  *HandlerLib
}

// handles event
// bool: whether event been handled
// error: handle error
func (c *Controller) HandleEvent(e *Event) (bool, error) {
	// see if we didn't handle it before
	c.mu.Lock()
	defer c.mu.Unlock()

	key := e.Name + e.Value
	_, found := c.ProcessedQ.Get(key)
	if found {
		return false, nil
	}

	h, hf := c.HandlersQ.Get(e.Name)
	if !hf {
		return false, fmt.Errorf("event %s not found", e.Name)
	}
	handleErr := h(e)
	return true, handleErr
}

func NewController() *Controller {
	c := new(Controller)
	c.ProcessedQ = NewEventQ()
	c.HandlersQ = NewHandlerLib()

	return c
}
