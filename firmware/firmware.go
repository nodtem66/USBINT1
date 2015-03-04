package firmware

import (
	"github.com/kylelemons/gousb/usb"
)

type FirmwareId int

// List of all support firmware
// Define your firmware here
const (
	FIRMWARE_TEMPERATURE_EP3_INT64 FirmwareId = iota
)

type Firmware interface {
	Open()
	Close()
	IOLoop()
	SetSender()
}

func GetFirmwareId(device usb.Device) FirmwareId {
	return 0
}
