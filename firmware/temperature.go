// Firmware for temperature sensor Interrupt 64 bytes EP3 0x83
// Length in byte     |  1   | 1  |         62          |
// Packet description | Temp | Id | 1 | 1 | 1 | ... | 1 |
package firmware

func runFirmwareTemperature(in chan []byte, out chan SqliteData) {
	for data := range in {
		out <- SqliteData{data[1], data[0]}
	}
}
