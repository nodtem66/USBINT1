package main

import (
	. "github.com/nodtem66/usbint1"
	. "github.com/nodtem66/usbint1/db"
	"testing"
)

func TestMain_CommandLine(t *testing.T) {
	// Test suit 1
	c := &CommandLine{
		PatientId: "T123",
		VidPid:    "1-1",
	}
	if err := c.ParseOption(); err == nil {
		t.Fatalf("This is not valid ", c)
	}
	// Test suit 2
	c.VidPid = "11"
	if err := c.ParseOption(); err == nil {
		t.Fatalf("This is not valid ", c)
	}
	// Test suit 3
	c.VidPid = "1:1"
	if err := c.ParseOption(); err != nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 4
	if c.Vid != 1 || c.Pid != 1 {
		t.Fatal("Error ParseOption for ", c)
	}
	// Test suit 5
	c.VidPid = "1000:00C3"
	if err := c.ParseOption(); err != nil {
		t.Fatal("This is not valid ", c)
	}
	if c.Vid != 0x1000 || c.Pid != 0x00C3 {
		t.Fatal("Error ParseOption for ", c)
	}
	// Test suit 6
	c.PatientId = "123"
	if err := c.ParseOption(); err != nil {
		t.Fatal("This is not valid ", c)
	}

	// Check PatientId only /[0-9a-zA-Z](1:)/
	// Test suit 7
	c.PatientId = "_aasdasdc"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 8
	c.PatientId = "weqw_q"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 9
	c.PatientId = "test:qwe"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 10
	c.PatientId = "qweP*s"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 11
	c.PatientId = "we%"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 12
	c.PatientId = "as#"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 13
	c.PatientId = "as$"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 14
	c.PatientId = "as@"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 15
	c.PatientId = "asdasd!"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 16
	c.PatientId = "poo^"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 17
	c.PatientId = "test&"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 18
	c.PatientId = "asdasd("
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 19
	c.PatientId = "asdasd)"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 20
	c.PatientId = "asdasd-asd"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 21
	c.PatientId = "aqwe=we"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 22
	c.PatientId = "qwee;qwe"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 23
	c.PatientId = "wqe.wewe"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 24
	c.PatientId = "qwe,e"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
	// Test suit 25
	c.PatientId = "ewqwe/qwe"
	if err := c.ParseOption(); err == nil {
		t.Fatal("This is not valid ", c)
	}
}

func TestMain_Integrate(t *testing.T) {
	io := NewIOHandle()
	io.Dev.OpenDevice(0x10C4, 0x8846)
	if io.Dev.OpenErr != nil {
		t.Fatal(io.Dev.OpenErr)
	}

	sqlite := NewSqliteHandle()
	sqlite.PatientId = "T100"
	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
}
