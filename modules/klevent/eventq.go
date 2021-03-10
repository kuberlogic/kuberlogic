package klevent

type EventQ struct {
	Q map[string]*Event
}

func (q *EventQ) Add(key string, e *Event) {
	q.Q[key] = e
}

func (q *EventQ) Get(key string) (*Event, bool) {
	e, f := q.Q[key]
	return e, f
}

func (q *EventQ) Del(key string) {
	delete(q.Q, key)
}

func NewEventQ() *EventQ {
	return &EventQ{
		Q: make(map[string]*Event, 0),
	}
}
