/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package klevent

import (
	"fmt"
	"sync"
)

const (
	eventNameField  = "kuberlogic.com/event-name"
	eventTypeField  = "kuberlogic.com/event-type"
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

	h, hf := c.HandlersQ.Get(e.Type)
	if !hf {
		return false, fmt.Errorf("event %s type %s handler not found", e.Name, e.Type)
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
