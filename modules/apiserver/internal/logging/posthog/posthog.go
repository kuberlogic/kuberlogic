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
