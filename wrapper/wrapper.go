package wrapper

import (
	"fmt"
	. "github.com/nodtem66/usbint1/config"
	. "github.com/nodtem66/usbint1/event"
	. "github.com/nodtem66/usbint1/firmware"
	"sync/atomic"
)

type WrapperInitFunc func([]byte) error

var WrapperFuncMap = map[FirmwareId]WrapperInitFunc{
	FIRMWARE_TEMPERATURE_EP3_INT64: WrapperTemperatureEP3Int64,
}

func NewWrapper(fid FirmwareId, out chan struct{}, e *EventHandler) chan []byte {
	input := make(chan []byte, 4096)
	wrapperEvent := NewEventSubcriptor()
	e.Subcribe(EVENT_WRAPPER, wrapperEvent)

	var count uint64 = 0
	go func() {
		//TODO: wrap data from input channel
		for in := range input {

			err := WrapperFuncMap[fid](in)
			if err != nil {
				fmt.Printf("\nWrapper Error: %s\n", in)
			}
			atomic.AddUint64(&count, 1)
		}
	}()

	go func() {
		for {
			select {
			case msg := <-wrapperEvent.Pipe:
				if msg.Status == EVENT_MAIN_TO_EXIT {
					total_count := atomic.LoadUint64(&count)
					if DEBUG {
						if LOG_LEVEL >= 3 {
							fmt.Printf("\nstop wrapper loop for firmwareID(%d)"+
								"\ntotal_count(%d)\n", fid, total_count)
						}
						wrapperEvent.Done <- fmt.Sprintf("c%d", total_count)
					} else {
						wrapperEvent.Done <- struct{}{}
					}
					return
				}
			}
		}
	}()
	return input
}
