package firmware

import (
	"fmt"
	. "github.com/nodtem66/usbint1"
	. "github.com/nodtem66/usbint1/db"
	"strings"
	"time"
)

type Firmware struct {
	Id, Vendor, Product int
	Err                 error
	VendorString        string
	ProductString       string
}

func (t *Firmware) String() string {
	return fmt.Sprintf("Firmware(%d)@device(%04X:%04X)", t.Id, t.Vendor, t.Product)
}
func NewFirmware(io *IOHandle, sqlite *SqliteHandle) *Firmware {
	f := &Firmware{Err: fmt.Errorf("No Firmware")}

	if io.Dev == nil {
		return f
	}
	if io.Dev.OpenErr != nil {
		return f
	}
	vendor, err := io.Dev.Device.GetStringDescriptor(1)
	if err != nil {
		f.Err = err
		return f
	}
	product, err := io.Dev.Device.GetStringDescriptor(2)
	if err != nil {
		f.Err = err
		return f
	}

	// set vid pid
	f.Vendor = int(io.Dev.Device.Vendor)
	f.Product = int(io.Dev.Device.Product)
	f.VendorString = vendor
	f.ProductString = product

	// init firmware
	if strings.HasPrefix(vendor, "Silicon Laboratories Inc.") &&
		strings.HasPrefix(product, "Fake Streaming 64byt") {
		f.Id = 1
		// setting Tag depended on USB device
		sqlite.Unit = "Celcius"
		sqlite.ReferenceMin = 0
		sqlite.ReferenceMax = 100
		sqlite.Resolution = 100
		sqlite.SamplingRate = time.Millisecond

		// create new tag
		f.Err = sqlite.EnableMeasurement([]string{"id", "temperature"})
	} else if strings.HasPrefix(vendor, "CardioArt") &&
		strings.Contains(product, "oximeter") {
		f.Id = 2
		sqlite.Unit = "mV"
		sqlite.ReferenceMin = 0
		sqlite.ReferenceMax = 3.3
		sqlite.Resolution = 4194304 //22 bit
		sqlite.SamplingRate = time.Millisecond

		// create new tag
		f.Err = sqlite.EnableMeasurement([]string{"LED2", "LED1"})
	} else if strings.HasPrefix(vendor, "CardioArt") &&
		strings.Contains(strings.ToLower(product), "ecg") {
		f.Id = 3
		sqlite.Unit = "mV"
		sqlite.ReferenceMin = 0
		sqlite.ReferenceMax = 3.3
		sqlite.Resolution = 16777216 //22 bit
		sqlite.SamplingRate = time.Millisecond

		// create new tag
		f.Err = sqlite.EnableMeasurement([]string{"Lead-I", "Lead-II", "Lead-III"})
	}
	if f.Id == 0 {
		f.Err = fmt.Errorf("No Firmware")
	}
	return f
}
