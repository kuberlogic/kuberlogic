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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Event struct {
	Name  string
	Type  string
	Value string
}

func NewEvent(name, eType, val string) (*Event, error) {
	if name == "" || val == "" {
		return nil, fmt.Errorf("event name or value can't be empty")
	}
	return &Event{
		Name:  name,
		Type:  eType,
		Value: val,
	}, nil
}

func NewEventMeta(meta *v1.ObjectMeta) (*Event, bool) {
	n, _ := meta.Annotations[eventNameField]
	v, _ := meta.Annotations[eventValueField]
	t, _ := meta.Annotations[eventTypeField]

	ev, err := NewEvent(n, t, v)

	return ev, err == nil
}

func RegisterEventMeta(meta v1.ObjectMeta, e *Event) {
	meta.Annotations[eventNameField] = e.Name
	meta.Annotations[eventValueField] = e.Value
	meta.Annotations[eventTypeField] = e.Type
}
