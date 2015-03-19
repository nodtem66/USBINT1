package wrapper

import (
	"fmt"
	. "github.com/nodtem66/usbint1/event"
	. "github.com/nodtem66/usbint1/firmware"
)

type WrapperInitFunc func([]byte) error

var WrapperFuncMap = map[FirmwareId]WrapperInitFunc{
	FIRMWARE_TEMPERATURE_EP3_INT64: WrapperTemperatureEP3Int64,
}

func NewWrapper(fid FirmwareId, out chan struct{}, e *EventHandler) chan []byte {
	input := make(chan []byte, 100)
	wrapperEvent := NewEventSubcriptor()
	e.Subcribe(EVENT_WRAPPER, wrapperEvent)

	go func() {
		//TODO: wrap data from input channel
		count := 0
	channel_loop:
		for in := range input {
			select {
			case msg := <-wrapperEvent.Pipe:
				if msg.Status == EVENT_MAIN_TO_EXIT {
					fmt.Printf("\nstop wrapper loop for firmwareID(%d)"+
						"\ntotal_count(%d)\n", fid, count)
					break channel_loop
				}
			default:

				err := WrapperFuncMap[fid](in)
				if err != nil {
					fmt.Printf("\nWrapper Error: %s\n", in)
				}
				count++
			}
		}
	}()
	return input
}
