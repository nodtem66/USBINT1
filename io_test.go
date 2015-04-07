package usbint

import (
	"testing"
	"time"
)

func TestIO_StartStop(t *testing.T) {
	io := NewIOHandle()
	io.Start()
	io.Stop()
}
func Test_Libusb(t *testing.T) {
	io := NewIOHandle()
	io.Dev.OpenDevice(0x10C4, 0x8846)
	if io.Dev == nil {
		t.Fatal("No device")
	}
	t.Logf("Endpoint 0x%02X [%d]\n", io.Dev.EpAddr, io.Dev.maxSize)
	ep, err := io.Dev.Device.OpenEndpoint(1, 0, 0, 0x83)
	if err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 64)
	_, err = ep.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Temperature: %d", buffer[0])
}

func TestIO_OpenDevice(t *testing.T) {
	io := NewIOHandle()

	io.Dev.OpenDevice(0x10C4, 0x8846)
	if io.Dev == nil {
		t.Fatal("No Open device")
	} else {
		t.Logf("Endpoint 0x%02X [%d]\n", io.Dev.EpAddr, io.Dev.maxSize)
	}

	// null loop
	go func() {
		for _ = range io.Pipe {
		}
	}()

	io.Start()
	defer io.Stop()
	time.Sleep(time.Second)

}
