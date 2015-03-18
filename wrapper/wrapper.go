package wrapper

import (
	"fmt"
	. "github.com/nodtem66/usbint1/firmware"
)

type WrapperInitFunc func([]byte) error

var WrapperFuncMap = map[FirmwareId]WrapperInitFunc{
	FIRMWARE_TEMPERATURE_EP3_INT64: WrapperTemperatureEP3Int64,
}

func NewWrapper(fid FirmwareId, out chan struct{}, done chan struct{}) chan []byte {
	input := make(chan []byte, 100)

	go func() {
		//TODO: wrap data from input channel
		count := 0
	channel_loop:
		for in := range input {
			select {
			case <-done:
				fmt.Println("stop wrapper loop")
				break channel_loop
			default:

				err := WrapperFuncMap[fid](in)
				if err != nil {
					fmt.Printf("Wrapper Error: %s\n", in)
				}
				count++
			}
		}
	}()
	return input
}
