package main

import (
	"bufio"
	"fmt"
	"github.com/kylelemons/gousb/usb"
	"os"
)

func main() {
	// Only one context should be needed for an application.  It should always be closed.
	ctx := usb.NewContext()
	defer ctx.Close()

	// ListDevices is used to find the devices to open.
	devs, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		// The usbid package can be used to print out human readable information.
		fmt.Printf("%03d.%03d %s:%s\n", desc.Bus, desc.Address, desc.Vendor, desc.Product)
		if desc.Vendor == 0x10C4 && desc.Product == 0x8846 {
			return true
		}
		return false
	})

	// All Devices returned from ListDevices must be closed.
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	// ListDevices can occaionally fail, so be sure to check its return value.
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Found %d device\n", len(devs))
	if len(devs) > 0 {
		dev := devs[0]
		reader := bufio.NewReader(os.Stdin)
		for {
			ep, err := dev.OpenEndpoint(1, 0, 0, 0x81)
			if err != nil {
				fmt.Println(err)
				return
			}
			buffer := make([]byte, 8)
			length, err := ep.Read(buffer)
			if err != nil {
				fmt.Println("[err] ", err)
				return
			}
			fmt.Println("[ok] ", buffer[:length])
			fmt.Printf("$ ")
			text, _, _ := reader.ReadLine()
			if string(text) == "exit" {
				return
			}
		}
	}

}
