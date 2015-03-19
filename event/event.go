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
)

const (
	EVENT_MAIN_TO_EXIT EventDataType = iota
	EVENT_MAIN_EXITED
	EVENT_SCANNER_EXITED
	EVENT_IOLOOP_EXITED
	EVENT_WRAPPER_EXITED
	EVENT_DATABASE_EXITED
	EVENT__END
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
	Done chan struct{}
}

// InputPipe is a local channel receied data and sent to subcriptor
type inputPipe struct {
	pipe     chan EventMessage
	isClosed bool
}

func NewEventSubcriptor() *EventSubscriptor {
	e := &EventSubscriptor{
		Pipe: make(chan EventMessage, 1),
		Done: make(chan struct{}, 1),
	}
	return e
}

// EventHandler centralizes the in-event and distributes out-event
type EventHandler struct {
	subcriped_service map[EventName]*EventSubscriptor
	input_pipe        *inputPipe
	done              chan struct{}
}

func NewEventHandler() *EventHandler {
	handler := new(EventHandler)
	handler.subcriped_service = make(map[EventName]*EventSubscriptor)
	handler.input_pipe = &inputPipe{
		pipe:     make(chan EventMessage, 1),
		isClosed: false,
	}
	handler.done = make(chan struct{})
	return handler
}

func (handler EventHandler) Start() {
	go func() {

		for {
			select {
			case msg := <-handler.input_pipe.pipe:
				for name, service := range handler.subcriped_service {
					if msg.Name == EVENT_ALL || msg.Name == name {
						service.Pipe <- msg
					}
				}
			case <-handler.done:
				handler.input_pipe.isClosed = true
				for _, service := range handler.subcriped_service {
					close(service.Pipe)
				}
				//fmt.Println("exit loop ")
				//fmt.Printf("isClose %#v\n", handler.input_pipe.isClosed)
				return
			}
		}

	}()
}

func (handler EventHandler) Stop() chan struct{} {

	handler.done <- struct{}{}
	done := make(chan struct{})

	go func() {
		// wait for done signal for every subcripted service
		for name, ch := range handler.subcriped_service {
			if name != EVENT_ALL && name != EVENT_MAIN {
				//fmt.Printf("wait for %d\n", name)
				<-ch.Done
				//fmt.Printf("ok for %d\n", name)
			}
		}
		done <- struct{}{}
	}()
	return done
}

func (handler EventHandler) SendMessage(name EventName, msg EventDataType) {
	message := EventMessage{name, msg}
	//fmt.Printf("isClose %#v\n", handler.input_pipe.isClosed)
	if !handler.input_pipe.isClosed {
		//fmt.Printf("%d -> %d\n", name, msg)
		handler.input_pipe.pipe <- message
	}
}
func (handler EventHandler) Subcribe(name EventName, channel *EventSubscriptor) {
	if _, isNotEmpty := handler.subcriped_service[name]; isNotEmpty {
		delete(handler.subcriped_service, name)
	}
	handler.subcriped_service[name] = channel
}
