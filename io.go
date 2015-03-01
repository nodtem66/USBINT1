package usbint

import ()

type Scanner struct {
}

func (s Scanner) StartScan() {
	timeout := make(chan struct{})
	for {
		go scan_running()
		<-timeout
	}

}

func (s Scanner) StopScan() {

}

func scan_running() {
	for {
		//TODO: Implement scanner
	}

	//TODO: Implement sender
	for {
	}
}
