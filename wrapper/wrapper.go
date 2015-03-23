package wrapper

import (
	"fmt"
	. "github.com/nodtem66/usbint1/config"
	. "github.com/nodtem66/usbint1/db"
	. "github.com/nodtem66/usbint1/event"
	"sync/atomic"
)

type WrapperId int
type WrapperInitFunc func([]byte) ([]InfluxData, error)

const (
	WRAPPER_TEMPERATURE_SIMPLE WrapperId = iota
)

var WrapperFuncMap = map[WrapperId]WrapperInitFunc{
	WRAPPER_TEMPERATURE_SIMPLE: WrapperTemperatureSimple,
}

func NewWrapper(id WrapperId, e *EventHandler, out chan []InfluxData) chan []byte {
	input := make(chan []byte, 64)

	wrapperEvent := NewEventSubcriptor()
	e.Subcribe(EVENT_WRAPPER, wrapperEvent)

	var count uint64 = 0
	go func() {

		// process the data from firmware channel
		for in := range input {

			// wrap the data with wrapper function
			data, err := WrapperFuncMap[id](in)
			if err != nil {
				fmt.Printf("\nWrapper Error: %s\n", in)
			}

			atomic.AddUint64(&count, 1)

			// send to influxdb
			out <- data

		}
	}()

	go func() {
		for msg := range wrapperEvent.Pipe {
			if msg.Name == EVENT_WRAPPER && msg.Status == EVENT_WRAPPER_TO_EXIT {

				total_count := atomic.LoadUint64(&count)
				if DEBUG {
					if LOG_LEVEL >= 3 {
						fmt.Printf("\nstop wrapper loop for firmwareID(%d)"+
							"\ntotal_count(%d)\n", id, total_count)
					}
					wrapperEvent.Done <- fmt.Sprintf("c%d", total_count)
				} else {
					wrapperEvent.Done <- struct{}{}
				}
				close(input)
				e.SendMessage(EVENT_DATABASE, EVENT_DATABASE_TO_EXIT)

				return
			}
		}
	}()
	return input
}
