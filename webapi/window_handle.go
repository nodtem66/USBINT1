// +build windows

package webapi

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// check the status process true is running, false in dead
// processName is not include their extension
// for window input processName without exe
// this function warp tasklist.exe for list process
func IsProcessRunning(processName string) (isRun bool, err error) {
	isRun = false
	searchPattern := fmt.Sprintf(`Imagename eq %s.exe`, processName)
	cmd := exec.Command(`tasklist`, `/fi`, searchPattern, `/nh`, `/fo`, `CSV`)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	// execute command
	if err = cmd.Run(); err != nil {
		return
	}
	cmdString := cmdOutput.String()
	if !strings.Contains(strings.ToLower(cmdString), "no task") {
		if len(strings.Split(cmdString, ",")) > 1 {
			isRun = true
			return
		}
	}
	return
}

// list the pid from process name
// this function warp tasklist.exe for list process
func ListPidFromName(processName string) (pids []string, err error) {
	searchPattern := fmt.Sprintf(`Imagename eq %s.exe`, processName)
	cmd := exec.Command(`tasklist`, `/fi`, searchPattern, `/nh`, `/fo`, `CSV`)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	// execute command
	if err = cmd.Run(); err != nil {
		return
	}
	lines := strings.Split(cmdOutput.String(), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) > 2 {
			pids = append(pids, strings.Replace(fields[1], `"`, ``, -1))
		}
	}
	return
}

// check the input pid is running
func IsPIDRunning(pid string) (isRun bool, err error) {
	isRun = false
	searchPattern := fmt.Sprintf(`pid eq %s`, pid)
	cmd := exec.Command(`tasklist`, `/fi`, searchPattern, `/nh`, `/fo`, `CSV`)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	//execute command
	if err = cmd.Run(); err != nil {
		return
	}
	cmdString := cmdOutput.String()
	if !strings.Contains(strings.ToLower(cmdString), "no task") {
		if len(strings.Split(cmdString, ",")) > 1 {
			isRun = true
			return
		}
	}
	return
}

// this function warp nssm.exe for start `usbint` service
// i is index from record
// bus is device bus from usb descriptor
// addr is device address from usb descriptor
func StartUsbIntService(i int, patient string, bus int, addr int) (err error) {
	//nssm.exe set usbint1_%%i AppParameters ...
	serviceName := fmt.Sprintf("usbint1_%d", i)
	parameters := fmt.Sprintf("%s %d %d", patient, bus, addr)
	cmd := exec.Command(`nssm`, `set`, serviceName, `AppParameters`, parameters)
	log.Printf("nssm set %s AppParameters %s", serviceName, parameters)
	if err = cmd.Run(); err != nil {
		return
	}
	cmd = exec.Command(`nssm`, `start`, serviceName)
	log.Printf("nssm start %s", serviceName)
	err = cmd.Run()
	return
}

// this function warp nssm.exe for stop `usbint` service
func StopUsbIntService(i int) (err error) {
	serviceName := fmt.Sprintf("usbint1_%d", i)

	cmd := exec.Command(`nssm`, `stop`, serviceName)

	log.Printf("nssm stop %s", serviceName)
	err = cmd.Run()
	return
}