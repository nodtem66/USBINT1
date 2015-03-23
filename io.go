package usbint

import (
	"fmt"
	"github.com/kylelemons/gousb/usb"
	. "github.com/nodtem66/usbint1/config"
	. "github.com/nodtem66/usbint1/db"
	. "github.com/nodtem66/usbint1/event"
	. "github.com/nodtem66/usbint1/firmware"
	"time"
)

type ScannerStatus int

const (
	SCANNER_WAIT ScannerStatus = iota
	SCANNER_FOUND
	SCANNER_CONNECTED
)

const (
	EVENT_SCANNER_TO_RETRY EventDataType = iota
	EVENT_SCANNER_WAIT
	EVENT_SCANNER_FOUND
	EVENT_SCANNER_CONNECT
)

type Scanner struct {
	Context             *usb.Context
	VendorId, ProductId int
	Status              ScannerStatus
	EventChannel        *EventSubscriptor
	Done                chan struct{}
	Retry               chan struct{}
}

func NewScanner(vid, pid int) *Scanner {

	c := usb.NewContext()
	c.Debug(0)

	scanner := &Scanner{
		Context:      c,
		VendorId:     vid,
		ProductId:    pid,
		EventChannel: NewEventSubcriptor(),
		Status:       SCANNER_WAIT,
		Done:         make(chan struct{}, 3),
		Retry:        make(chan struct{}, 1),
	}

	return scanner
}

func (s *Scanner) StartScan(e *EventHandler, influx *InfluxHandle) {

	e.Subcribe(EVENT_SCANNER, s.EventChannel)

	// main routine
	go func() {
	start_scan_loop:
		for {
			// select all device with specific vid,pid
			e.SendMessage(EVENT_MAIN, EVENT_SCANNER_WAIT)
			devices, err := s.Context.ListDevices(func(desc *usb.Descriptor) bool {

				// check if device has a selected vid, pid
				if int(desc.Vendor) == s.VendorId && int(desc.Product) == s.ProductId {
					return true
				}
				return false
			})
			if err != nil && len(devices) == 0 {
				fmt.Println(err)
			}

			// select the first device that can be initialized
			var f Firmware
			for i, d := range devices {
				if i == 0 {
					f = NewFirmware(d)
					s.Status = SCANNER_FOUND
					e.SendMessage(EVENT_MAIN, EVENT_SCANNER_FOUND)
				} else {
					d.Close()
				}
			}

			// start firmware reader, else wait for retry
			if f != nil {
				if DEBUG && LOG_LEVEL >= 3 {
					fmt.Printf("Start %s\n", f)
				}
				e.SendMessage(EVENT_MAIN, EVENT_SCANNER_CONNECT)

				// run routine usb reader
				err := f.IOLoop(e, influx)
				if err != nil {
					fmt.Println(err)
					s.Retry <- struct{}{}
				}
			} else {
				fmt.Println("wait for device")
			}

			// wait for retry
			for {
				select {
				case <-time.After(time.Second * 3):
					if s.Status == SCANNER_WAIT {

						fmt.Println("timeout 3 second.")

						continue start_scan_loop
					}
				case <-s.Retry:
					fmt.Println("Retry!")
					continue start_scan_loop
				case <-s.Done:
					return
				}
			}
		}
	}()

	// manage external event handler
	go func() {
		for msg := range s.EventChannel.Pipe {
			if msg.Name == EVENT_ALL {
				switch msg.Status {
				case EVENT_SCANNER_TO_EXIT:
					fallthrough
				case EVENT_MAIN_TO_EXIT:
					s.StopScan()
					s.EventChannel.Done <- struct{}{}
					e.SendMessage(EVENT_IOLOOP, EVENT_IOLOOP_TO_EXIT)
				}
			} else {
				switch msg.Status {
				case EVENT_SCANNER_TO_RETRY:
					s.Retry <- struct{}{}
				}
			}
		}
	}()
}
func (s *Scanner) StopScan() {
	s.Done <- struct{}{}
}

func (s *Scanner) Close() {
	s.Context.Close()
}
