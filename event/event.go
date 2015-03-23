package event

type EventDataType int
type EventName int

const (
	EVENT_ALL EventName = iota
	EVENT_MAIN
	EVENT_SCANNER
	EVENT_IOLOOP
	EVENT_WRAPPER
	EVENT_DATABASE
	event_self
)

const (
	EVENT_MAIN_TO_EXIT EventDataType = iota
	EVENT_SCANNER_TO_EXIT
	EVENT_IOLOOP_TO_EXIT
	EVENT_WRAPPER_TO_EXIT
	EVENT_DATABASE_TO_EXIT
	event_self_destroy
)

// EventMessage structure of message in Event channel pipe
type EventMessage struct {
	Name   EventName
	Status EventDataType
}

// EventSubscriptor is a read-only channel to recieve EventMessage from
// EventHandler
type EventSubscriptor struct {
	Pipe chan EventMessage
	Done chan interface{}
}

// InputPipe is a local channel receied data and sent to subcriptor
type inputPipe struct {
	pipe     chan EventMessage
	isClosed bool
}

func NewEventSubcriptor() *EventSubscriptor {
	e := &EventSubscriptor{
		Pipe: make(chan EventMessage, 1),
		Done: make(chan interface{}, 1),
	}
	return e
}

// EventHandler centralizes the in-event and distributes out-event
type EventHandler struct {
	subcriped_service map[EventName]*EventSubscriptor
	input_pipe        *inputPipe
}

func NewEventHandler() *EventHandler {
	handler := new(EventHandler)
	handler.subcriped_service = make(map[EventName]*EventSubscriptor)
	handler.input_pipe = &inputPipe{
		pipe:     make(chan EventMessage, 1),
		isClosed: false,
	}

	return handler
}

func (handler *EventHandler) Start() {

	// start main rountine for incoming message
	go func() {
		for {
			select {
			case msg := <-handler.input_pipe.pipe:

				// check for self destroying message
				if msg.Name == event_self && msg.Status == event_self_destroy {

					// swap current pipe with null pipe
					handler.NullPipe()

					// close all subcriped channel
					/*
						for _, service := range handler.subcriped_service {
							close(service.Pipe)
						}
					*/
					return
				}
				// otherwise boardcast message to subcriped channel
				for name, service := range handler.subcriped_service {

					if msg.Name == EVENT_ALL || msg.Name == name {
						service.Pipe <- msg
					}

				}
			}
		}

	}()
}

func (handler *EventHandler) Stop() chan []interface{} {

	done := make(chan []interface{}, 1)

	// run shutdown routine
	go func() {
		returnValue := make([]interface{}, 0)
		// wait for done signal for every subcripted service
		for name, ch := range handler.subcriped_service {
			if name != EVENT_ALL && name != EVENT_MAIN {

				value := <-ch.Done
				returnValue = append(returnValue, value)
				close(ch.Pipe)

			}

		}
		// destroy the main routine
		handler.SendMessage(event_self, event_self_destroy)
		done <- returnValue
	}()

	return done
}

func (handler *EventHandler) SendMessage(name EventName, msg EventDataType) {

	message := EventMessage{name, msg}
	handler.input_pipe.pipe <- message

}
func (handler *EventHandler) Subcribe(name EventName, channel *EventSubscriptor) {
	if _, isNotEmpty := handler.subcriped_service[name]; isNotEmpty {
		delete(handler.subcriped_service, name)
	}
	handler.subcriped_service[name] = channel
}
func (handler *EventHandler) NullPipe() {
	// buffer routine for left incoming data
	go func() {
		for _ = range handler.input_pipe.pipe {
		}
	}()
}
