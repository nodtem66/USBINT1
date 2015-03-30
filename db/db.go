package db

const (
	LENGTH_QUEUE     = 1024
	TIMEOUT_MSEC     = 500
	TIMEOUT_USB_MSEC = 100
)

type DataTag struct {
	PatientId  string
	SignalType string
	Unit       string
	Resolution int
	Reference  float64
}

type DatabaseInterface interface {
	Start()
	Stop()
}
