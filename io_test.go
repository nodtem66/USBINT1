package usbint

import (
	"testing"
	"time"
)

func TestIO_StartStop(t *testing.T) {
	io := NewIOHandle()
	io.Start()
	time.Sleep(time.Millisecond * 500)
	io.Stop()
}
func Test_Libusb(t *testing.T) {
	io := NewIOHandle()

	io.OpenDevice(0x10C4, 0x8846)
	if io.Dev.Device == nil {
		t.Fatal(io.Dev.openErr)
	}
	t.Logf("Endpoint 0x%02X [%d]\n", io.Dev.EpAddr, io.Dev.maxSize)
	_, err := io.Dev.Device.OpenEndpoint(1, 0, 0, 0x83)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIO_OpenDevice(t *testing.T) {
	io := NewIOHandle()

	io.OpenDevice(0x10C4, 0x8846)
	if io.Dev.Device == nil {
		t.Fatal(io.Dev.openErr)
	}
	t.Logf("Endpoint 0x%02X [%d]\n", io.Dev.EpAddr, io.Dev.maxSize)

	go func() {
		l := 0
		for packet := range io.Pipe {
			l += int(packet[0])
		}
		t.Logf("[l = %d]", l)
	}()

	io.Start()
	//time.Sleep(time.Second)
	io.Stop()
}
