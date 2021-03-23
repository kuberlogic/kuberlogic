## General
`Kuberlogic Event` aka `klevent` library is a simple library for events handling in Kubernetes environment.

## Usage
### Event
`klevent` is based on two objects - `event` and `eventController`.

Controller is responsible for events storage and events processing, while `event` represents actual event.

Starting controller:
```go
// some code ommited
ctrl := klevent.NewController()
```

... and let's register some event:
```go
e, err := klevent.NewEvent("someEventName", "reallyBadEventType", "eventSupportingData")
// err handling
```

... we can also register event from Kubernetes annotations:
```go
// meta is v1.ObjectMeta object
e, err := klevent.NewEventMeta(meta)
// err handling
```

Finally, let's handle the event:
```go
handled, err := ctrl.Handle(e)
// handled is true if event handler is found and triggered
// err handling
```

### Event Handlers
Events need their handlers in order to be remediated.
```go
type HandlerFunc func(e *Event) error
```

Event handlers need to be registered in controller before they can be used:
```go
err := ctrl.RegisterHandlerFunc(handlerFunc, "usefulHandler") // later all events with usefulHandler event type will trigger this handler
if err != nil {
	// handler is already registered for this name
	ctrl.DeregisterHandlerFunc("usefulHandler")
	ctrl.RegisterHandlerFunc(handlerFunc, "usefulHandler")
}
```