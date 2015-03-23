package firmware

import (
	"github.com/kylelemons/gousb/usb"
	"github.com/nodtem66/usbint1/db"
	"github.com/nodtem66/usbint1/event"
)

type FirmwareId int
type FirmwareAcceptFunc func(string, string, *usb.Descriptor) bool
type FirmwareInitFunc func(*usb.Device) Firmware

// List of all support firmware
// Define your firmware with following step
// 1. define your firmware name `$NAME' after FIRMWARE_$YOURNAME
//    this is your firmware id `$ID'
// 2. define your struct in new .go file within the `firmware' package
//    this struct have to implement Firmware interface
// 3. define the device selection function in FirmwareAcceptFuncMap
//    this is a map $ID to your function name
// 4. define initial function in FirmwareInitFuncMap
//    this is a map $ID to your function name
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
	// This IOLoop will run under scanner with a selected device.
	// Inside this function, you freely code any possible task:
	// from open usb with your prefer endpoint, select the wrapper
	// and send to database
	// NOTE: this function have to start loop with goroutine
	IOLoop(*event.EventHandler, *db.InfluxHandle) error
	// return $ID for open wrapper interface
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
