package db

import "time"

const (
	LENGTH_QUEUE     = 1024
	TIMEOUT_MSEC     = 500
	TIMEOUT_USB_MSEC = 100
)

type DataTag struct {
	IdTag        int64
	PatientId    string
	NumChannel   int
	Measurement  string
	Unit         string
	Resolution   int
	ReferenceMin float64
	ReferenceMax float64
	Descriptor   []string
	SamplingRate time.Duration
}

type DatabaseInterface interface {
	Start()
	Stop()
}
