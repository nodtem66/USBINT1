package db

import (
	"fmt"
	"github.com/influxdb/influxdb/client"
	. "github.com/nodtem66/usbint1/event"
	"net/url"
	"strconv"
	"time"
)

const (
	DEFAULT_HOST     = "localhost"
	DEFAULT_PORT     = 8086
	DEFAULT_DATABASE = "dev"
	LENGTH_QUEUE     = 1024
	TIMEOUT_MSEC     = 500
	TIMEOUT_USB_MSEC = 100
)

// Infomation for a patientid at a specific time
// eg. {"channel1": 1001, "channel2": 1201.1}
// {"unit": "mV", "resolution": 12, "reference": 5}
type InfluxData map[string]interface{}

// Type for inter-channel exchange. this worker encode into BatchPoint
// depened on the version of influxdb
/*
type InfluxData struct {
	Timestamp time.Time // This timestamp is not used
	Data      []InfluxDataMap
}*/

// A Database worker for insert streaming data into influxdb
// usage:
// * NewInflux() to initial the worker
// * NewInfluxWithHostPort(host, port) to initial worker with custom config
// * SetUserPassword(user, pass) to set the auth to worker. can be change after
//    initialized worker
// * Start()/Stop() to manage the worker
type InfluxHandle struct {
	Client       *client.Client
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	Version      string
	PatientId    string
	SignalType   string
	Unit         string
	Resolution   int
	SamplingTime time.Duration
	Timestamp    time.Time
	Reference    float64
	Pipe         chan []InfluxData
	EventChannel *EventSubscriptor
	Done         chan struct{}
	ShouldDump   bool
}

// new the influx object with a default_host(localhost), default_port(8086),
// username(root), and password(root)
func NewInflux() *InfluxHandle {
	influx := &InfluxHandle{
		Host:         DEFAULT_HOST,
		Port:         DEFAULT_PORT,
		Username:     "root",
		Password:     "root",
		Pipe:         make(chan []InfluxData, LENGTH_QUEUE),
		EventChannel: NewEventSubcriptor(),
		Done:         make(chan struct{}, 1),
		Resolution:   1,
		Reference:    1,
		PatientId:    "none",
	}
	return influx
}

// new influx client with host (string) and port (int)
func NewInfluxWithHostPort(host string, port int) *InfluxHandle {
	influx := NewInflux()
	influx.Host = host
	influx.Port = port
	return influx
}

// set user and password to influx client
func (i *InfluxHandle) SetUserPassword(username string, password string) {
	if len(username) != 0 {
		i.Username = username
	}
	if len(password) != 0 {
		i.Password = password
	}
}

func (i *InfluxHandle) SetSignalType(signal string) {
	i.SignalType = signal
}

func (i *InfluxHandle) SetPatientId(id string) {
	if len(id) > 0 {
		i.PatientId = id
	}
}

func (i *InfluxHandle) SetUnit(unit string) {
	i.Unit = unit
}

func (i *InfluxHandle) SetResolution(r int) {
	if r > 0 {
		i.Resolution = r
	}
}
func (i *InfluxHandle) SetReference(r float64) {
	i.Reference = r
}
func (i *InfluxHandle) SetSamplingTime(t time.Duration) {
	i.SamplingTime = t
}

// create the connection and check connectivity
func (i *InfluxHandle) Connect() error {
	var cl *client.Client

	// setting url
	u := url.URL{Scheme: "http"}
	u.Host = fmt.Sprintf("%s:%d", i.Host, i.Port)
	u.User = url.UserPassword(i.Username, i.Password)

	// new connection to influx
	cl, err := client.NewClient(
		client.Config{
			URL:       u,
			Username:  i.Username,
			Password:  i.Password,
			UserAgent: "Usbint firmware",
		})
	// if there is error
	if err != nil {
		return err
	}

	// ping to server, testing the connection
	i.Client = cl
	if _, v, e := i.Client.Ping(); e != nil {
		//fmt.Printf("failed to connect to %s\n", i.Client.Addr())
		return e
	} else {
		i.Version = v
		if !i.ShouldDump {
			//fmt.Printf("Connected to %s version %s\n", i.Client.Addr(), i.Version)
		}
	}
	return nil
}

// execute the command with in specific database
func (i *InfluxHandle) Query(query string, database string) (*client.Results, error) {
	result, err := i.Client.Query(client.Query{Command: query, Database: database})
	if err != nil {
		return nil, err
	}

	if result.Error() != nil {
		return nil, result.Error()
	}
	return result, nil
}

func (i *InfluxHandle) Start(e *EventHandler) {
	e.Subcribe(EVENT_DATABASE, i.EventChannel)

	go func() {
	connection_loop:
		for {
			// try to connection
			err := i.Connect()
			if err != nil {
				fmt.Println(err)
			} else {

				buffer := make([]InfluxData, 0)
				// process incoming queue channel
				for data := range i.Pipe {
					// check for delay data more than TIMEOUT_USN_MSEC msec
					// update the global timestamp

					for _, v := range data {
						// store the point in buffer
						buffer = append(buffer, v)
					}

					// release 100 point
					isTimeout := time.Now().Sub(i.Timestamp) > time.Duration(TIMEOUT_USB_MSEC)*time.Millisecond
					if isTimeout {
						if len(buffer) > 0 {
							err := i.send(buffer)
							if err != nil {
								fmt.Printf("%s\n", err)
							}
							buffer = buffer[:0]
						}
						i.Timestamp = time.Now()
					}
					if len(buffer) > 100 {
						err := i.send(buffer)
						if err != nil {
							fmt.Printf("%s\n", err)
						}
						buffer = buffer[:0]
					} /*
						select {
						case <-i.Done:
							return
						}*/
				}
			}
			// wait for retry
			for {
				select {
				case <-time.After(time.Millisecond * TIMEOUT_MSEC):
					continue connection_loop
				case <-i.Done:
					return
				}
			}
		}
	}()

	go func() {
		for msg := range i.EventChannel.Pipe {
			if msg.Name == EVENT_DATABASE && msg.Status == EVENT_DATABASE_TO_EXIT {
				i.Done <- struct{}{}
				i.EventChannel.Done <- struct{}{}
			}
		}
	}()

}

func (i *InfluxHandle) Stop() {
	i.Done <- struct{}{}
}

// transform the InfluxData to BatchPoint and send
func (ifx *InfluxHandle) send(data []InfluxData) error {
	var bp client.BatchPoints
	var nanosecTime time.Time
	points := make([]client.Point, len(data))

	bp.Database = DEFAULT_DATABASE
	bp.Precision = "ms"
	nanosecTime = ifx.Timestamp
	nanosecTime = client.SetPrecision(ifx.Timestamp, "ms")

	// cache tags parameters for speed
	tags := map[string]string{
		"type":       ifx.SignalType,
		"resolution": strconv.Itoa(ifx.Resolution),
		"reference":  strconv.FormatFloat(ifx.Reference, 'f', 2, 64),
		"unit":       ifx.Unit,
	}

	for i, value := range data {
		points[i].Name = ifx.PatientId
		points[i].Tags = tags
		points[i].Timestamp = nanosecTime
		points[i].Fields = (map[string]interface{})(value)
		nanosecTime = nanosecTime.Add(ifx.SamplingTime)
	}
	ifx.Timestamp = nanosecTime
	bp.Points = points

	results, err := ifx.Client.Write(bp)
	if err != nil {
		return err
	}
	if results != nil && results.Err != nil {
		return results.Err
	}
	return nil
}

// insert InfluxData into sending queue
func (ifx *InfluxHandle) Send(data []InfluxData) {
	ifx.Pipe <- data
}

func (i *InfluxHandle) nullPipe() {
	go func() {
		for _ = range i.Pipe {
		}
	}()
}
