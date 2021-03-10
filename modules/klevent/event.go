package klevent

import (
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Event struct {
	Name  string
	Value string
}

func NewEvent(name, val string) (*Event, error) {
	if name == "" || val == "" {
		return nil, fmt.Errorf("event name or value can't be empty")
	}
	return &Event{
		Name:  name,
		Value: val,
	}, nil
}

func NewEventMeta(meta *v1.ObjectMeta) (*Event, bool) {
	n, _ := meta.Annotations[eventNameField]
	v, _ := meta.Annotations[eventValueField]

	ev, err := NewEvent(n, v)

	return ev, err == nil
}

func RegisterEventMeta(meta v1.ObjectMeta, e *Event) {
	meta.Annotations[eventNameField] = e.Name
	meta.Annotations[eventValueField] = e.Value
}
