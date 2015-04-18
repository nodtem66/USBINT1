package main

import (
	"fmt"
	"github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
	"github.com/peterh/liner"
	"os"
	"strconv"
	"strings"
)

var (
	isRunning          = true
	history_file       = ".liner_history"
	commands_completes = []string{"exit", "dev", "ep", "close", "help", "name"}
)

type Command struct {
	usbContext *usb.Context
	usbDev     *usb.Device
	line       *liner.State
}

func (c *Command) OpenDevice(vid usb.ID, pid usb.ID, num int, isVerbose bool) error {
	// ListDevices is used to find the devices to open.
	devs, err := c.usbContext.ListDevices(func(desc *usb.Descriptor) bool {
		// The usbid package can be used to print out human readable information.
		if isVerbose {
			fmt.Printf("[%03d.%03d] %s:%s %s %s\n",
				desc.Bus, desc.Address, desc.Vendor, desc.Product,
				usbid.Describe(desc), usbid.Classify(desc))
		}
		if desc.Vendor == vid && desc.Product == pid {
			return true
		}
		return false
	})
	// clear lib: libusb not found [code -5] error
	if err == usb.ERROR_NOT_FOUND {
		err = nil
	}
	// ListDevices can occaionally fail, so be sure to check its return value.
	if err != nil {
		return err
	}
	if len(devs) == 0 {
		return fmt.Errorf("No devices")
	}

	// All Devices returned from ListDevices must be closed.
	for i, d := range devs {
		if i != num {
			d.Close()
		} else {
			c.usbDev = devs[num]
		}
	}
	return nil
}
func (c *Command) CloseDevice() {
	if c.usbDev != nil {
		c.usbDev.Close()
		c.usbDev = nil
	}
}
func (c *Command) Close() {
	c.CloseDevice()
	if c.usbContext != nil {
		c.usbContext.Close()
	}
}
func (c *Command) ParseCommand(raw string) (err error) {
	line := strings.Split(raw, " ")
	switch line[0] {
	case "exit":
		isRunning = false
	case "help":
		fmt.Println("[ok] Help\n exit\t\t\tquit the program\n",
			"dev\t\t\tlist the device\n dev <vid> <pid>\topen device for vid and pid\n",
			"close\t\t\tclose current opened device\n ep\t\t\tlist all endpoints for",
			"current device\n ep <address> <length>\tread <length> bytes from <address>")
	case "dev":
		if len(line) == 1 {
			fmt.Println("[ok] List all devices:")
			err = c.OpenDevice(0, 0, 0, true)

		} else {
			if len(line) >= 3 {
				var vid, pid, num int64
				if vid, err = strconv.ParseInt(line[1], 16, 32); err != nil {
					return
				}
				if pid, err = strconv.ParseInt(line[2], 16, 32); err != nil {
					return
				}
				fmt.Printf("[ok] Open device %02X:%02X\n", vid, pid)
				if len(line) >= 4 {
					if num, err = strconv.ParseInt(line[3], 0, 32); err != nil {
						return
					}
					err = c.OpenDevice(usb.ID(vid), usb.ID(pid), int(num), false)
				} else {
					err = c.OpenDevice(usb.ID(vid), usb.ID(pid), 0, false)
				}

			} else {
				err = fmt.Errorf("Invalid params")
			}
		}
	case "close":
		c.CloseDevice()
	case "get":
		if len(line) == 1 {
			err = fmt.Errorf("Invalid params")
			return
		}
		switch line[1] {
		case "string":
			if len(line) < 3 {
				err = fmt.Errorf("Missing params")
			}
			var index int64
			if index, err = strconv.ParseInt(line[2], 0, 32); err == nil {
				var mydesc string
				if mydesc, err = c.usbDev.GetStringDescriptor(int(index)); err == nil {
					fmt.Println("[ok] ", mydesc)
				}
			}
		}
	case "ep":
		if c.usbDev == nil {
			err = fmt.Errorf("No device")
			return
		}
		if len(line) == 1 {
			fmt.Println("[ok] List Endpoint in Config[1] Interface[0] Setup[0]")
			endpoints := c.usbDev.Configs[0].Interfaces[0].Setups[0].Endpoints
			for _, e := range endpoints {
				fmt.Printf("[%d] %s\n", e.Address, e.String())
			}
		} else if len(line) < 3 {
			err = fmt.Errorf("Invalid params")
			return
		} else {
			var ep, maxLen uint64
			var length int
			var endp usb.Endpoint
			var text string
			// converion string to int
			if ep, err = strconv.ParseUint(line[1], 0, 8); err != nil {
				return
			}
			if maxLen, err = strconv.ParseUint(line[2], 0, 64); err != nil {
				return
			}

			// create buffer reader and endpoint
			buffer := make([]byte, maxLen)
			if endp, err = c.usbDev.OpenEndpoint(1, 0, 0, uint8(ep)); err != nil {
				return
			}
		loop_read_endpoint:
			for {
				if length, err = endp.Read(buffer); err != nil {
					return
				}
				fmt.Println("[ok] ", buffer[:length])
				// get input commands
				if text, err = c.line.Prompt(""); err != nil {
					return
				}
				if len(text) > 0 {
					break loop_read_endpoint
				}
			}
		}
	default:
		err = fmt.Errorf("Invalid command")
	}
	return
}
func main() {
	// initial local commnad
	cmd := &Command{usbContext: usb.NewContext()}
	defer cmd.Close()

	// initial Liner for unix-like shell
	line := liner.NewLiner()
	cmd.line = line
	defer line.Close()
	defer func() {
		if f, err := os.Create(history_file); err != nil {
			fmt.Println("[err] writing history file: ", err)
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}()
	// set command line autocompletion
	line.SetCompleter(func(line string) (c []string) {
		for _, n := range commands_completes {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})
	line.SetCtrlCAborts(true)

	// load history file
	if f, err := os.Open(history_file); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	// main routine
	fmt.Println("usbtest v0.1.0")
	for isRunning {
		// get input command from user
		if text, err := line.Prompt("> "); err != nil {
			fmt.Println("\n[err] readline ", err)
			if err == liner.ErrPromptAborted {
				isRunning = false
			}
		} else {
			// Parse command board
			if err := cmd.ParseCommand(text); err != nil {
				fmt.Println("[err]", err)
			}
			line.AppendHistory(text)
		}
	}
	/*
		if len(devs) > 0 {

			// liner
			dev := devs[0]
			reader := bufio.NewReader(os.Stdin)
			for {
				// get input commands
				fmt.Printf("$ ")
				text, _, _ := reader.ReadLine()
				if string(text) == "exit" {
					return
				}
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

			}
		}
	*/
}
