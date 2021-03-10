package klevent

type HandlerFunc func(e *Event) error

type HandlerLib struct {
	Q map[string]HandlerFunc
}

func (l *HandlerLib) Get(key string) (HandlerFunc, bool) {
	h, f := l.Q[key]
	return h, f
}

func (l *HandlerLib) RegisterHandlerFunc(h HandlerFunc, name string) {
	l.Q[name] = h
}

func NewHandlerLib() *HandlerLib {
	return &HandlerLib{
		Q: make(map[string]HandlerFunc, 0),
	}
}
