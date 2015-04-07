// USBINT Firmware for host
package usbint

import (
	"fmt"
	"github.com/kylelemons/gousb/usb"
	"time"
)

const (
	LENGTH_PIPE = 1024
)

type IOHandle struct {
	StartTime  time.Time
	PacketRead int64
	PacketPipe int64
	Quit       chan bool
	Pipe       chan []byte
	Dev        *DeviceHandle
}
type DeviceHandle struct {
	Context  *usb.Context
	Device   *usb.Device
	Endpoint usb.Endpoint
	maxSize  int
	EpAddr   int
	OpenErr  error
}

func NewIOHandle() *IOHandle {
	io := &IOHandle{
		Quit: make(chan bool),
		Pipe: make(chan []byte, LENGTH_PIPE),
		Dev:  &DeviceHandle{OpenErr: usb.ERROR_NO_DEVICE},
	}
	return io
}
func (i *DeviceHandle) OpenDevice(vid, pid int) {

	i.Context = usb.NewContext()
	i.Context.Debug(3)
	// scan for usb device that match vid pid
	var devices []*usb.Device
	devices, i.OpenErr = i.Context.ListDevices(func(desc *usb.Descriptor) bool {
		if desc.Vendor == usb.ID(vid) && desc.Product == usb.ID(pid) {
			return true
		}
		return false
	})
	if len(devices) == 0 {
		i.OpenErr = fmt.Errorf("No devices")
	}

	// clear lib: libusb not found [code -3] error
	if i.OpenErr == usb.ERROR_NOT_FOUND {
		i.OpenErr = nil
	}
	// if Err, close handle
	if i.OpenErr != nil {
		i.Context.Close()
		return
	}
	// in the case devices have same VID/PID; the first openable device is selected
	for i, dev := range devices {
		if i != 0 {
			dev.Close()
		}
	}
	i.Device = devices[0]
	eps := devices[0].Configs[0].Interfaces[0].Setups[0].Endpoints
	ep := eps[len(eps)-1]
	i.maxSize = int(ep.MaxPacketSize)
	i.EpAddr = int(ep.Address)
	i.Endpoint, i.OpenErr = i.Device.OpenEndpoint(1, 0, 0, ep.Address)
}
func (i *IOHandle) Start() {

	// start timer
	fmt.Printf("[IO Start at %s]\n", time.Now())
	i.StartTime = time.Now()

	// main routine
	if i.Dev != nil && i.Dev.OpenErr == nil {
		go i.runReader(i.Dev.Endpoint)
	} else {
		go i.runWaiter()
	}
}

func (i *IOHandle) Stop() {
	// send shutdown signal
	i.Quit <- true
	<-i.Quit
	// stop timer
	totalSec := time.Now().Sub(i.StartTime).Seconds()
	fmt.Printf("[IO Stop] [Running Time %f sec] [Read %d Loss %d]\n", totalSec, i.PacketRead, i.PacketRead-i.PacketPipe)
	// check the device is opened
	if i.Dev != nil && i.Dev.OpenErr == nil {
		// then close all connection
		i.Dev.Device.Close()
		i.Dev.Context.Close()
	}
}

func (i *IOHandle) runReader(endpoint usb.Endpoint) {
	// main loop
	for i.Pipe != nil {

		// prepare buffer
		buffer := make([]byte, i.Dev.maxSize)

		// read
		_, err := endpoint.Read(buffer)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		i.PacketRead++

		select {
		case i.Pipe <- buffer:
			i.PacketPipe++
		case <-i.Quit:
			i.Quit <- true
			return
		}

	}
}

func (i *IOHandle) runWaiter() {
	<-i.Quit
	i.Quit <- true
}

func isOpen(ch chan bool) (ok bool) {
	_, ok = <-ch
	return
}
