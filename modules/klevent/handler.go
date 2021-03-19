package klevent

import "fmt"

type HandlerFunc func(e *Event) error

type HandlerLib struct {
	Q map[string]HandlerFunc
}

func (l *HandlerLib) Get(key string) (HandlerFunc, bool) {
	h, f := l.Q[key]
	return h, f
}

func (l *HandlerLib) RegisterHandlerFunc(h HandlerFunc, name string) error {
	if _, found := l.Q[name]; found {
		return fmt.Errorf("handler with %s name is already registerd", name)
	}
	l.Q[name] = h
	return nil
}

func (l *HandlerLib) DeregisterHandlerFunc(name string) {
	delete(l.Q, name)
}

func NewHandlerLib() *HandlerLib {
	return &HandlerLib{
		Q: make(map[string]HandlerFunc, 0),
	}
}
