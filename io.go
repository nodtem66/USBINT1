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
	Context *usb.Context
	Device  *usb.Device
	maxSize int
	EpAddr  int
	openErr error
}

func NewIOHandle() *IOHandle {
	io := &IOHandle{
		Quit: make(chan bool),
		Pipe: make(chan []byte, LENGTH_PIPE),
	}
	return io
}
func (i *IOHandle) OpenDevice(vid, pid int) error {
	// scan for usb device that match vid pid
	i.Dev = &DeviceHandle{
		Context: usb.NewContext(),
	}
	var devices []*usb.Device
	devices, i.Dev.openErr = i.Dev.Context.ListDevices(func(desc *usb.Descriptor) bool {
		if desc.Vendor == usb.ID(vid) && desc.Product == usb.ID(pid) {
			return true
		}
		return false
	})
	if len(devices) == 0 {
		i.Dev.openErr = fmt.Errorf("No devices")
	}
	// if Err, close handle
	if i.Dev.openErr != nil && i.Dev.openErr != usb.ERROR_NOT_FOUND {
		i.Dev.Context.Close()
		return i.Dev.openErr
	}
	// in the case devices have same VID/PID; the first openable device is selected
	for i, dev := range devices {
		if i != 1 {
			dev.Close()
		}
	}
	i.Dev.Device = devices[0]
	eps := devices[0].Configs[0].Interfaces[0].Setups[0].Endpoints
	ep := eps[len(eps)-1]
	i.Dev.maxSize = int(ep.MaxPacketSize)
	i.Dev.EpAddr = int(ep.Address)
	return nil
}
func (i *IOHandle) Start() {

	// start timer
	fmt.Printf("[IO Start at %s]\n", time.Now())
	i.StartTime = time.Now()
	// main routine
	go func() {

		// Pre-initialize variable
		var length int
		var err, openErr error
		var buffer []byte
		var endpoint usb.Endpoint
		isOpen := false

		// check the device is opened
		if i.Dev != nil && i.Dev.openErr == nil {
			// prepare buffer
			buffer = make([]byte, i.Dev.maxSize)
			// open endpoint
			endpoint, openErr = i.Dev.Device.OpenEndpoint(1, 1, 1, uint8(i.Dev.EpAddr))
			if openErr != nil {
				fmt.Printf("Error: %s\n", openErr)
			} else {
				isOpen = true
			}
		}

		// main loop
		for {
			// read
			if isOpen {
				length, err = endpoint.Read(buffer)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
				} else {
					i.PacketRead++
					select {
					case i.Pipe <- buffer[:length]:
						i.PacketPipe++
					}
				}
			}

			select {
			case <-i.Quit:
				// check the device is opened
				if i.Dev != nil && i.Dev.openErr == nil {
					// then close all connection
					i.Dev.Context.Close()
					i.Dev.Device.Close()
				}
				i.Quit <- true
				return
			}
		}
	}()
	return
}

func (i *IOHandle) Stop() {
	// send shutdown signal
	i.Quit <- true
	// wait for shutdown process
	<-i.Quit
	// stop timer
	totalSec := time.Now().Sub(i.StartTime).Seconds()
	fmt.Printf("[IO Stop] [Running Time %f sec] [Read %d Loss %d]\n", totalSec, i.PacketRead, i.PacketRead-i.PacketPipe)
}

func isOpen(ch chan bool) (ok bool) {
	_, ok = <-ch
	return
}
