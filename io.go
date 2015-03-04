package usbint

import (
	"fmt"
	"github.com/kylelemons/gousb/usb"
)

type Scanner struct {
	context             *usb.Context
	vendorId, productId int
	timeoutMsec         int
	eventChannel        chan EventMessage
}

func NewScanner(vid, pid int) *Scanner {

	c := usb.NewContext()
	c.Debug(0)

	scanner := &Scanner{
		context:      c,
		vendorId:     vid,
		productId:    pid,
		eventChannel: make(chan EventMessage, 3),
	}
	return scanner
}

func (s Scanner) StartScan() error {
	// select all device with specific vid,pid
	devices, err := s.context.ListDevices(func(desc *usb.Descriptor) bool {

		// check if device has a selected vid, pid
		if int(desc.Vendor) == s.vendorId && int(desc.Product) == s.productId {
			return true
		}
		return false
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, d := range devices {
		d.Close()
	}
	return nil
}

func (s Scanner) StopScan() {

}

func (s Scanner) Close() {
	s.context.Close()
}

func scan_running() {
	for {
		//TODO: Implement scanner
	}

	//TODO: Implement sender
	for {
	}
}

type Device struct {
	*usb.Device
	ManufacturerString string
	ProductString      string
}

func NewDeivce(dev *usb.Device) *Device {
	device := &Device{
		dev,
		"",
		"",
	}
	return device
}
