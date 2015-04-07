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
func NewFirmware(io IOHandle, sql SqliteHandle) (f *Firmware) {
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

		// setting firmware infomation
		f.Id = 1

		// setting Tag depended on USB device
		sqlite.PatientId = c.PatientId
		sqlite.Unit = "Celcius"
		sqlite.ReferenceMin = 0
		sqlite.ReferenceMax = 100
		sqlite.Resolution = 100
		sqlite.SamplingRate = time.Millisecond

		// create new tag
		if err := sqlite.EnableMeasurement([]string{"id", "temperature"}); err != nil {
			f.Err = err
		}
	}
	return
}

func (f *Firmware) Start() {
	//TODO: separate firmware
	switch f.Id {
	case 1:
		go runFirmwareTemperature(f.InPipe, f.OutPipe)
	default:
		go runNoFirmware()
	}
}

func (f *Firmware) Stop() {
	f.Quit <- true
	<-f.Quit
}

func runNoFirmware() {
	<-f.Quit
	f.Quit <- true
}
