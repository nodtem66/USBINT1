package firmware

import (
	"fmt"
	"github.com/kylelemons/gousb/usb"
	. "github.com/nodtem66/usbint1/event"
	"strings"
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

func (t *TemperatureEP3Int64) IOLoop(w chan []byte, event *EventHandler) error {
	endpoint, err := t.OpenEndpoint(1, 0, 0, 0x83)
	if err != nil {
		return err
	}

	// subcribe event from global event manager
	firmwareEvent := NewEventSubcriptor()
	event.Subcribe(EVENT_IOLOOP, firmwareEvent)

	go func() {
		for {
			buffer := make([]byte, 64)
			length, err := endpoint.Read(buffer)
			if err != nil {
				fmt.Println(err)
				return
			}
			select {
			case w <- buffer[:length]:
			case msg := <-firmwareEvent.Pipe:
				if msg.Status == EVENT_MAIN_TO_EXIT {
					fmt.Printf("\nclose IOLoop for %s\n", t)
					close(w)
					return
				}
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
