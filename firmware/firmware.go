package firmware

import (
	"github.com/kylelemons/gousb/usb"
)

type FirmwareId int
type FirmwareAcceptFunc func(string, string, *usb.Descriptor) bool
type FirmwareInitFunc func(*usb.Device) Firmware

// List of all support firmware
// Define your firmware here
const (
	FIRMWARE_NO_FOUND FirmwareId = iota
	FIRMWARE_TEMPERATURE_EP3_INT64
)

var FirmwareAcceptFuncMap = map[FirmwareId]FirmwareAcceptFunc{
	FIRMWARE_TEMPERATURE_EP3_INT64: TemperatureEP3Int64AcceptFunc,
}

var FirmwareInitFuncMap = map[FirmwareId]FirmwareInitFunc{
	FIRMWARE_TEMPERATURE_EP3_INT64: TemperatureEP3Int64InitFunc,
}

type Firmware interface {
	IOLoop(chan []byte, chan struct{}) error
	GetFirmwareId() FirmwareId
}

func NewFirmware(dev *usb.Device) Firmware {

	vendor, _ := dev.GetStringDescriptor(1)
	product, _ := dev.GetStringDescriptor(2)
	desc := dev.Descriptor

	for id, accept := range FirmwareAcceptFuncMap {
		if accept(vendor, product, desc) {
			return FirmwareInitFuncMap[id](dev)
		}
	}
	return nil
}
