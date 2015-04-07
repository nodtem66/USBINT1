package firmware

import (
	"fmt"
	. "github.com/nodtem66/usbint1"
	. "github.com/nodtem66/usbint1/db"
	"strings"
)

type Firmware struct {
	Id, Vendor, Product int
	Err                 error
	InPipe              chan []byte
	OutPipe             chan SqliteData
	Quit                chan bool
}

func (t *Firmware) String() string {
	return fmt.Sprintf("Firmware(%d)@device(%04X:%04X)", t.Id, t.Vendor, t.Product)
}
func NewFirmware(patientId string, io *IOHandle, sql *SqliteHandle) error {
	//f = &Firmware{Err: fmt.Errorf("No Firmware")}
	//f.Quit = make(chan bool)
	if io.Dev == nil {
		return fmt.Errorf("No Firmware")
	}
	vendor, err := io.Dev.Device.GetStringDescriptor(1)
	if err != nil {
		//f.Err = err
		return err
	}
	product, err := io.Dev.Device.GetStringDescriptor(2)
	if err != nil {
		//f.Err = err
		return err
	}

	// set pipe
	f.InPipe = io.Pipe
	f.OutPipe = sql.Pipe

	// set vid pid
	f.Vendor = int(io.Dev.Device.Vendor)
	f.Product = int(io.Dev.Device.Product)

	// init firware
	err := fmt.Errorf("Cannot init firmware")
	if strings.HasPrefix(vendor, "Silicon Laboratories Inc.") &&
		strings.HasPrefix(product, "Fake Streaming 64byt") {
		f.Id = 1
		err = initFirmwareTemperature(patientId, sql)
	}
	return err
}
