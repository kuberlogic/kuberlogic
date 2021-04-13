package posthog

import (
	ph "github.com/posthog/posthog-go"
	"os"
)

type Event struct {
	event string
	props map[string]interface{}
}

type postHog struct {
	client   ph.Client
	hostname string
}

var postHogInstance postHog

func Init(apiKey string) (ph.Client, error) {
	client := ph.New(apiKey)
	postHogInstance.client = client
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	postHogInstance.hostname = hostname
	return client, nil
}

func NewMessage(event string) *Event {
	return &Event{
		event: event,
		props: make(map[string]interface{}),
	}
}

func (p *Event) With(key string, value interface{}) *Event {
	p.props[key] = value
	return p
}

func (p *Event) Create() error {
	if postHogInstance.client == nil {
		return nil // make nothing if the posthog is not initialized
	}
	return postHogInstance.client.Enqueue(ph.Capture{
		DistinctId: postHogInstance.hostname,
		Event:      p.event,
		Properties: p.props,
	})
}
