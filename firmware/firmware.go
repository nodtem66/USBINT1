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
func NewFirmware(patientId string, io *IOHandle, sql *SqliteHandle) (f *Firmware) {
	f.Err = fmt.Errorf("No Firmware")
	f.Quit = make(chan bool)
	if io.Dev == nil {
		return
	}
	vendor, err := io.Dev.Device.GetStringDescriptor(1)
	if err != nil {
		f.Err = err
		return
	}
	product, err := io.Dev.Device.GetStringDescriptor(2)
	if err != nil {
		f.Err = err
		return
	}

	// set pipe
	f.InPipe = io.Pipe
	f.OutPipe = sql.Pipe

	// set vid pid
	f.Vendor = int(io.Dev.Device.Vendor)
	f.Product = int(io.Dev.Device.Product)

	// init firware
	if strings.HasPrefix(vendor, "Silicon Laboratories Inc.") &&
		strings.HasPrefix(product, "Fake Streaming 64byt") {
		f.Id = 1
	}
	f.initFirmware(patientId, sql)
	return
}

func (f *Firmware) initFirmware(patientId string, sql *SqliteHandle) {
	switch f.Id {
	case 1:
		f.Err = initFirmwareTemperature(patientId, sql)
	}
}

func (f *Firmware) Start() {
	//TODO: separate firmware
	switch f.Id {
	case 1:
		go runFirmwareTemperature(f.InPipe, f.OutPipe)
	default:
		go runNoFirmware(f)
	}
}

func (f *Firmware) Stop() {
	f.Quit <- true
	<-f.Quit
}

func runNoFirmware(f *Firmware) {
	<-f.Quit
	f.Quit <- true
}
