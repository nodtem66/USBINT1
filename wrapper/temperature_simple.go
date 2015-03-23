package wrapper

import (
	. "github.com/nodtem66/usbint1/db"
)

func WrapperTemperatureSimple(buffer []byte) (data *InfluxData, err error) {

	data = &InfluxData{
		Data: []InfluxDataMap{
			InfluxDataMap{"counter": buffer[1], "channel1": buffer[0]},
		}}
	//fmt.Printf("%d:%d ", buffer[0], buffer[1])
	return
}
