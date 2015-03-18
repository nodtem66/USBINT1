package usbint

import (
	"testing"
)

func TestEvent_Start(t *testing.T) {
	e := NewEventHandler()
	e.Start()
	e.Stop()
}

func TestEvent_Subscriptor(t *testing.T) {
	ch := make(EventSubscriptor, 10)

	e := NewEventHandler()
	e.Start()
	e.Subcribe(1, ch)

	if _, ok := e.subcribed_service[1]; !ok {
		t.Fail()
	}

	e.Stop()
}

func TestEvent_SendMessage(t *testing.T) {
	e := NewEventHandler()
	e.Start()
	e.SendMessage(0, 10)
	e.Stop()
}

func TestEvent_SendMessageToSubscriptor(t *testing.T) {
	ch := make(EventSubscriptor, 10)

	e := NewEventHandler()
	e.Start()
	e.Subcribe(1, ch)

	go func() {
		e.SendMessage(1, 10)
		e.SendMessage(1, 11)
		e.SendMessage(1, 12)
		e.Stop()
	}()

	for msg := range ch {
		t.Logf("(%d) %d\n", msg.Name, msg.Status)
	}
}

func TestEvent_SendMessageToSubscriptor2(t *testing.T) {
	ch := make(EventSubscriptor, 10)

	e := NewEventHandler()
	e.Start()
	e.Subcribe(1, ch)

	go func() {
		for msg := range ch {
			t.Logf("(%d) %d\n", msg.Name, msg.Status)
		}
	}()
	e.SendMessage(1, 13)
	e.SendMessage(1, 14)
	e.SendMessage(2, 10)
	e.Stop()
}
func TestEvent_SendMessageToGlobal(t *testing.T) {
	ch := make(EventSubscriptor)
	ch2 := make(EventSubscriptor)

	e := NewEventHandler()
	e.Start()
	e.Subcribe(1, ch)
	e.Subcribe(2, ch2)

	go func() {
		for msg := range ch {
			t.Logf("ch: (%d) %d\n", msg.Name, msg.Status)
		}
	}()
	go func() {
		for msg := range ch2 {
			t.Logf("ch2: (%d) %d\n", msg.Name, msg.Status)
		}
	}()

	e.SendMessage(1, 13)
	e.SendMessage(1, 14)
	e.SendMessage(2, 10)
	e.SendMessage(0, 10)
	e.Stop()
}
