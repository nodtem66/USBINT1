// +build windows

package webapi

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// check the status process true is running, false in dead
// processName is not include their extension
// for window input processName without exe
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

func StartProcess(execName string, args ...string) (cmd *exec.Cmd, out *bytes.Buffer, err error) {
	if len(strings.Split(execName, ".")) <= 1 {
		execName += ".exe"
	}
	cmd = exec.Command("START", "/B", "/MIN", execName, "5001")
	out = &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = out
	go func() {
		if err = cmd.Start(); err != nil {
			return
		}
		if err = cmd.Wait(); err != nil {
			return
		}
	}()

	return
}
