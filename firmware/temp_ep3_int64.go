package firmware

import (
	"fmt"
	"github.com/kylelemons/gousb/usb"
	. "github.com/nodtem66/usbint1/config"
	. "github.com/nodtem66/usbint1/db"
	. "github.com/nodtem66/usbint1/event"
	. "github.com/nodtem66/usbint1/wrapper"
	"strings"
	"time"
)

type TemperatureEP3Int64 struct {
	*usb.Device
}

func TemperatureEP3Int64AcceptFunc(
	vendor string,
	product string,
	desc *usb.Descriptor) bool {
	if strings.HasPrefix(vendor, "Silicon Laboratories Inc.") &&
		strings.HasPrefix(product, "Fake Streaming 64byt") {
		return true
	}
	return false
}

func TemperatureEP3Int64InitFunc(dev *usb.Device) Firmware {
	f := &TemperatureEP3Int64{
		dev,
	}
	return f
}

func (t *TemperatureEP3Int64) IOLoop(event *EventHandler, influx *InfluxHandle) error {

	// open usb device
	endpoint, err := t.OpenEndpoint(1, 0, 0, 0x83)
	if err != nil {
		return err
	}
	// InfluxDB config
	// you could set the parameter of data
	/*
		influx.SetPatientId("111#1")
		influx.SetReference(3.3)
		influx.SetResolution(1024)
		influx.SetSamplingTime(time.Millisecond)
		influx.SetSignalType("none")
		influx.SetUnit("d")
		influx.SetUserPassword("dev", "dev")
	*/

	// new wrapper
	w := NewWrapper(WRAPPER_TEMPERATURE_SIMPLE, event, influx.Pipe)

	// subcribe event from global event manager
	firmwareEvent := NewEventSubcriptor()
	event.Subcribe(EVENT_IOLOOP, firmwareEvent)

	// make exit channel for main routine
	done := make(chan struct{}, 1)

	// run main routine
	go func() {

		for {
			// read data from usb endpoint
			buffer := make([]byte, 64)
			length, err := endpoint.Read(buffer)
			if err != nil {
				fmt.Println(err)
				return
			}

			// wait for shutodown signal
			select {
			case w <- buffer[:length]:
			case <-done:
				event.SendMessage(EVENT_WRAPPER, EVENT_WRAPPER_TO_EXIT)
				return
			}
			// send to wrapper

		}
	}()

	// read internal message from event handler
	go func() {
		var timestamp int64
		if DEBUG {
			timestamp = time.Now().UnixNano()
		}
		for msg := range firmwareEvent.Pipe {
			// process internal message from event handler
			if msg.Name == EVENT_IOLOOP && msg.Status == EVENT_IOLOOP_TO_EXIT {

				// destroy main routine
				done <- struct{}{}

				if DEBUG {
					// set final total time in msec for benchmark program
					total_time_msec := (time.Now().UnixNano() - timestamp) / 1000000

					if LOG_LEVEL >= 3 {
						// output the shutdown message
						fmt.Printf("\nclose IOLoop for %s\ntotal_time(%.3f sec)\n", t,
							(float32)(total_time_msec)/1000)
					}
					firmwareEvent.Done <- fmt.Sprintf("t%d", total_time_msec)
				} else {
					firmwareEvent.Done <- struct{}{}
				}
				//close(w)
			}
		}
	}()
	return nil
}
func (t *TemperatureEP3Int64) GetFirmwareId() FirmwareId {
	return FIRMWARE_TEMPERATURE_EP3_INT64
}
func (t *TemperatureEP3Int64) String() string {
	return fmt.Sprintf("TemperatureEP3Int64 Firmware(%v)@device(%04X:%04X)", &t, int(t.Vendor), int(t.Product))
}
