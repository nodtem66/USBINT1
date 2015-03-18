package usbint

type EventDataType int
type EventName int

const (
	EVENT_ALL EventName = iota
	EVENT_MAIN
	EVENT_SCANNER
	EVENT_IOLOOP
	EVENT_SENDER
	EVENT_DATABASE
)

// EventMessage structure of message in Event channel pipe
type EventMessage struct {
	Name   EventName
	Status EventDataType
}

// EventHandler centralizes the in-event and distributes out-event
type EventHandler struct {
	subcribed_service map[EventName]EventSubscriptor
	input_pipe        chan EventMessage
	done              chan struct{}
}

// EventSubscriptor is a read-only channel to recieve EventMessage from
// EventHandler
type EventSubscriptor chan EventMessage

func NewEventHandler() *EventHandler {
	handler := new(EventHandler)
	handler.subcribed_service = make(map[EventName]EventSubscriptor)
	handler.input_pipe = make(chan EventMessage)
	handler.done = make(chan struct{})
	return handler
}

func (handler EventHandler) Start() {
	go func() {
		for {
			select {
			case msg := <-handler.input_pipe:
				for name, service := range handler.subcribed_service {
					if msg.Name == EVENT_ALL || msg.Name == name {
						service <- msg
					}
				}
			case <-handler.done:
				for _, service := range handler.subcribed_service {
					close(service)
				}
				return
			}
		}
	}()
}

func (handler EventHandler) Stop() {
	handler.done <- struct{}{}
}

func (handler EventHandler) SendMessage(name EventName, msg EventDataType) {
	message := EventMessage{name, msg}
	handler.input_pipe <- message
}
func (handler EventHandler) Subcribe(name EventName, channel EventSubscriptor) {
	if _, isNotEmpty := handler.subcribed_service[name]; isNotEmpty {
		delete(handler.subcribed_service, name)
	}
	handler.subcribed_service[name] = channel
}
