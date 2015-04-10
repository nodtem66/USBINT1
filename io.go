// USBINT Firmware for host
package usbint

import (
	"fmt"
	"github.com/kylelemons/gousb/usb"
	"time"
)

const (
	LENGTH_PIPE = 2048
)

type IOHandle struct {
	StartTime    time.Time
	PacketRead   int64
	PacketPipe   int64
	PacketErr    int64
	Quit         chan bool
	Pipe         chan []int64
	Dev          *DeviceHandle
	SamplingRate time.Duration
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
		Quit:         make(chan bool),
		Dev:          &DeviceHandle{OpenErr: usb.ERROR_NO_DEVICE},
		SamplingRate: time.Millisecond,
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

// SetPipe()
// set the sqlite input pipe
func (i *IOHandle) SetProperty(pipe chan []int64, d time.Duration) {
	i.Pipe = pipe
	i.SamplingRate = d
}

// Start()
// parameters:
//   id: firmware id to select the USB data byte
// return:
//   none
func (i *IOHandle) Start(id int) {

	// start timer
	fmt.Printf("[IO Start at %s]\n", time.Now())
	i.StartTime = time.Now()

	// main routine
	if i.Dev != nil && i.Dev.OpenErr == nil {
		go i.runReader(id)
	} else {
		go i.runWaiter()
	}
}

// Stop()
// stop the routine
func (i *IOHandle) Stop() {
	// send shutdown signal
	i.Quit <- true
	<-i.Quit
	// stop timer
	totalSec := time.Now().Sub(i.StartTime).Seconds()
	fmt.Printf("[IO Stop] [Running Time %f sec] [Read %d Loss %d Err %d]\n",
		totalSec, i.PacketRead, i.PacketRead-i.PacketPipe, i.PacketErr)
	// check the device is opened
	if i.Dev != nil && i.Dev.OpenErr == nil {
		// then close all connection
		i.Dev.Device.Close()
		i.Dev.Context.Close()
	}
}

func (i *IOHandle) runReader(id int) {

	// prepare buffer
	endpoint := i.Dev.Endpoint
	buffer := make([]byte, i.Dev.maxSize)
	// prepare timestamp
	var timestamp time.Time

	// main loop
main_loop:
	for i.Pipe != nil {

		// read
		_, err := endpoint.Read(buffer)
		if err != nil {
			//fmt.Printf("Error: %s\n", err)
			i.PacketErr++
			continue main_loop
		}
		i.PacketRead++
		isTimeout := time.Now().Sub(timestamp) > i.SamplingRate*5
		if isTimeout {
			timestamp = time.Now()
		}
		// parse buffer depended on firmware id
		var data []int64
		switch id {
		case 1:
			data = []int64{timestamp.UnixNano(), int64(buffer[1]), int64(buffer[0])}
		default:
			data = []int64{}
		}
		timestamp = timestamp.Add(i.SamplingRate)

		// send the data
		select {
		case i.Pipe <- data:
			//counting the sending data for checking data loss
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
