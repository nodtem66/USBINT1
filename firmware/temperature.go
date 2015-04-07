// Firmware for temperature sensor Interrupt 64 bytes EP3 0x83
// Length in byte     |  1   | 1  |         62          |
// Packet description | Temp | Id | 1 | 1 | 1 | ... | 1 |
package firmware

import (
	. "github.com/nodtem66/usbint1/db"
	"time"
)

func initFirmwareTemperature(patientId string, sqlite *SqliteHandle) (err error) {
	// setting Tag depended on USB device
	sqlite.PatientId = patientId
	sqlite.Unit = "Celcius"
	sqlite.ReferenceMin = 0
	sqlite.ReferenceMax = 100
	sqlite.Resolution = 100
	sqlite.SamplingRate = time.Millisecond

	// create new tag
	err = sqlite.EnableMeasurement([]string{"id", "temperature"})
	return
}
func runFirmwareTemperature(in chan []byte, out chan SqliteData) {
	for data := range in {
		out <- SqliteData{int64(data[1]), int64(data[0])}
	}
}
