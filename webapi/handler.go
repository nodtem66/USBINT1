package webapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/kylelemons/gousb/usb"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nodtem66/usbint1/config"
)

type Tag struct {
	Id           int    `json:"id"`
	Mnt          string `json:"mnt"`
	Unit         string `json:"unit"`
	Resolution   string `json:"resolution"`
	RefMin       string `json:"ref_min"`
	RefMax       string `json:"ref_max"`
	SamplingRate int64  `json:"sampling_rate"`
	Description  string `json:"description"`
	NumChannel   int    `json:"num_channel"`
	Active       bool   `json:"active"`
}
type Measurement map[string]interface{}

type MeasurementDescriptor struct {
	Name        string   `json:"name"`
	LastTime    int64    `json:"last_time"`
	TotalRecord int64    `json:"total_record"`
	ChannelName []string `json:"channel_name"`
}

func NewAPIRouter(h *APIHandler) *httprouter.Router {
	// config the router
	router := httprouter.New()
	router.GET("/version", h.HelpPage)
	router.GET("/patient", h.GetPatients)
	router.GET("/patient/:patientId/tag", h.GetTags)
	router.GET("/patient/:patientId/tag/:tagId", h.GetTag)
	router.GET("/patient/:patientId/mnt", h.GetMeasurements)
	router.GET("/patient/:patientId/mnt/:mntId", h.GetMeasurement)
	router.GET("/sys/:option", h.GetSystemStatus)
	router.NotFound = http.FileServer(http.Dir("app/")).ServeHTTP
	return router
}

//------------------------------------------------------------------------------
// Start APIHandler Section
//------------------------------------------------------------------------------

// APIHandler is a routine routed with api path
// conf store config from toml file
// version and commit is used for development version
// servicerecord is record about pid for each service
//   servicerecord[0] is pid of service 1
//   servicerecord[N] is pid of service N-1
// busrecord
type ServiceInfo struct {
	Pid       string
	PatientId string
	Bus       int
	Address   int
}
type APIHandler struct {
	Conf              *config.TomlConfig
	Version           string
	Commit            string
	ServiceRecord     []ServiceInfo
	StartedPid        map[string]bool
	StartedBusAddress map[string]bool
	SqliteConn        map[string]*sql.DB
}

// New APIHandler with version -1 and commit nil
func New() *APIHandler {
	h := &APIHandler{
		Version:           "-1",
		Commit:            "",
		ServiceRecord:     make([]ServiceInfo, 10),
		StartedPid:        map[string]bool{},
		StartedBusAddress: map[string]bool{},
		SqliteConn:        map[string]*sql.DB{},
	}
	return h
}

// singleton for cache SqliteConn
// use path filename for hashmap
func (h *APIHandler) GetSqliteConnection(path string) (conn *sql.DB, err error) {
	var ok bool
	if conn, ok = h.SqliteConn[path]; ok {
		return
	}
	if conn, err = sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=rw&_busy_timeout=5000", path)); err != nil {
		return
	}
	h.SqliteConn[path] = conn
	return
}

// Index page print program name and version
func (h *APIHandler) HelpPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "USBAPI [Version: %s] [Commit: %s]\n\n", h.Version, h.Commit)
	fmt.Fprintf(w, `### Api ###
/patient
/patient/:patientId/tag
/patient/:patientId/tag?query
/patient/:patientId/tag/:tagId
/patient/:patientId/mnt</li>
/patient/:patientId/mnt/:mntId
/patient/:patientId/mnt/:mntId?limit=100&ch=ch1,ch2&before=10s&after=11ns&orderby=asc`)
}

// Handler for /patient
// List of all patient db files in devices
func (h *APIHandler) GetPatients(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	patientId := []string{}
	ret := map[string]interface{}{}

	files, err := ioutil.ReadDir(h.Conf.DB.Path)
	if err != nil {
		log.Println(err)
	}
	for _, file := range files {
		if file.IsDir() == false {
			name := file.Name()
			if strings.HasSuffix(name, ".db") {
				patientId = append(patientId, strings.Replace(name, ".db", "", -1))
			}
		}
	}
	ret["result"] = patientId
	jsonRet, _ := json.Marshal(ret)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, string(jsonRet))
}

// Handler for
// 1. /patient/:patientId/tag
//   List of all tags in PatientId
// 2. /patient/:patientId/tag?query
//   List of all tags in PatientId with condition
func (h *APIHandler) GetTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err0 error
	ret := map[string]interface{}{}
	defer func() {
		if err0 != nil {
			ret["err"] = err0.Error()
		}
		jsonByte, _ := json.Marshal(ret)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, string(jsonByte))
	}()

	query := r.URL.Query()
	patientId := ps.ByName("patientId")

	// check valid patient id
	dbFileName := path.Join(h.Conf.DB.Path, patientId+".db")
	if _, err := os.Stat(dbFileName); err != nil {
		err0 = fmt.Errorf("no patient")
		return
	}
	var conn *sql.DB
	var err error
	if conn, err0 = h.GetSqliteConnection(dbFileName); err0 != nil {
		return
	}

	// no query string
	var rows *sql.Rows
	if len(query) == 0 {
		rows, err = conn.Query(`SELECT id,mnt,unit,resolution,ref_min,
			ref_max,sampling_rate,descriptor,num_channel,active FROM tag LIMIT 50`)
	} else { // if query string
		if _, ok := query["active"]; ok {
			rows, err = conn.Query(`SELECT id,mnt,unit,resolution,ref_min,
			ref_max,sampling_rate,descriptor,num_channel,active 
			FROM tag WHERE active=1 LIMIT 50`)
		} else if _, ok := query["inactive"]; ok {
			rows, err = conn.Query(`SELECT id,mnt,unit,resolution,ref_min,
			ref_max,sampling_rate,descriptor,num_channel,active
			FROM tag WHERE active=0 LIMIT 50`)
		}
	}
	// report query error
	if err != nil {
		err0 = err
		return
	}
	defer rows.Close()

	// enumerate db rows
	results := []Tag{}
	for rows.Next() {
		result := Tag{}
		rows.Scan(&result.Id, &result.Mnt, &result.Unit, &result.Resolution,
			&result.RefMin, &result.RefMax, &result.SamplingRate,
			&result.Description, &result.NumChannel, &result.Active)
		results = append(results, result)
	}

	// report rows error
	if rows.Err() != nil {
		err0 = rows.Err()
		return
	}
	ret["result"] = results
	return
}

// Handler for /patient/:patientId/tag/:tagId
// List of all descriptions in tagId
func (h *APIHandler) GetTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err0 error
	ret := map[string]interface{}{}
	defer func() {
		if err0 != nil {
			ret["err"] = err0.Error()
		}
		jsonByte, _ := json.Marshal(ret)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, string(jsonByte))
	}()

	patientId := ps.ByName("patientId")
	tagId := ps.ByName("tagId")

	// check valid patient id
	dbFileName := path.Join(h.Conf.DB.Path, patientId+".db")
	if _, err := os.Stat(dbFileName); err != nil {
		err0 = fmt.Errorf("no patient")
		return
	}
	var conn *sql.DB
	var err error
	if conn, err0 = h.GetSqliteConnection(dbFileName); err0 != nil {
		return
	}

	result := Tag{}
	if err = conn.QueryRow(
		`SELECT id,mnt,unit,resolution,ref_min,ref_max,sampling_rate,
		descriptor,num_channel,active FROM tag WHERE id = ?`, tagId).Scan(&result.Id,
		&result.Mnt, &result.Unit, &result.Resolution,
		&result.RefMin, &result.RefMax, &result.SamplingRate,
		&result.Description, &result.NumChannel, &result.Active); err != nil {
		err0 = fmt.Errorf("no tag")
		return
	}
	ret["result"] = result
	return
}

// Handler for /patient/:patientId/mnt
// List of all measurement units in patientId
func (h *APIHandler) GetMeasurements(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err0 error
	ret := map[string]interface{}{}
	defer func() {
		if err0 != nil {
			ret["err"] = err0.Error()
		}
		jsonByte, _ := json.Marshal(ret)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, string(jsonByte))
	}()

	patientId := ps.ByName("patientId")

	// check valid patient id
	dbFileName := path.Join(h.Conf.DB.Path, patientId+".db")
	if _, err := os.Stat(dbFileName); err != nil {
		err0 = fmt.Errorf("no patient")
		return
	}
	var conn *sql.DB
	var err error
	if conn, err0 = h.GetSqliteConnection(dbFileName); err0 != nil {
		return
	}

	// query table_name
	var rows *sql.Rows
	rows, err = conn.Query(`SELECT name FROM sqlite_master WHERE type='table';`)
	if err != nil {
		err0 = err
		return
	}

	names := []string{}
	for rows.Next() {
		var name string
		rows.Scan(&name)
		if name != "tag" {
			names = append(names, name)
		}
	}

	if rows.Err() != nil {
		err0 = err
		return
	}
	ret["result"] = names
	return
}

// Handler for
// 1. /patient/:patientId/mnt/:mntId
//   Description of measurement unit
// 2. /patient/:patientId/mnt/:mntId?query
//   List of signal from measurement unit with condition
func (h *APIHandler) GetMeasurement(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err0 error
	ret := map[string]interface{}{}
	defer func() {
		if err0 != nil {
			ret["err"] = err0.Error()
		}
		jsonByte, _ := json.Marshal(ret)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, string(jsonByte))
	}()

	query := r.URL.Query()
	patientId := ps.ByName("patientId")
	mntId := ps.ByName("mntId")

	// check mntId
	sp := strings.Split(mntId, "_")
	if len(sp) < 2 {
		err0 = fmt.Errorf("invalid measurement unit")
		return
	}

	var tagId int64
	if id, err := strconv.ParseInt(sp[len(sp)-1], 10, 64); err != nil {
		err0 = err
		return
	} else {
		tagId = int64(id)
	}

	// check valid patient id
	dbFileName := path.Join(h.Conf.DB.Path, patientId+".db")
	if _, err := os.Stat(dbFileName); err != nil {
		err0 = fmt.Errorf("no patient")
		return
	}
	var conn *sql.DB
	if conn, err0 = h.GetSqliteConnection(dbFileName); err0 != nil {
		return
	}

	// no query string
	if len(query) == 0 {
		// struct for measurement description
		desc := MeasurementDescriptor{Name: mntId}

		// query total record
		if err := conn.QueryRow(
			fmt.Sprintf(`SELECT count(_rowid_) as total FROM %s ORDER BY time DESC LIMIT 1;`, mntId),
		).Scan(&desc.TotalRecord); err != nil {
			err0 = err
			return
		}

		// query last time
		if desc.TotalRecord > 0 {
			if err := conn.QueryRow(
				fmt.Sprintf(`SELECT time FROM %s ORDER BY time DESC LIMIT 1;`, mntId),
			).Scan(&desc.LastTime); err != nil {
				err0 = err
				return
			}
		}

		// query channel_name
		var jsonDesc string
		if err := conn.QueryRow(
			`SELECT descriptor FROM tag WHERE id = ?`, tagId,
		).Scan(&jsonDesc); err != nil {
			err0 = err
			return
		}
		// return json format for measurement unit description
		json.Unmarshal([]byte(jsonDesc), &desc.ChannelName)
		ret["result"] = desc
		return
	} else {
		// with query string
		// default: ?limit=100&ch=0&desc
		//    query the lastest 100 records from channel 0
		// ?after=123ns (ASC)
		//    query the lastest 100 records after 123ns from channel 0
		// ?before=123ns (DESC)
		//	  query the lastest 100 record before 123ns
		// time unit: s ms us ns
		//    1 == 1s == 1000ms == 1000000us == 1000000000ns
		mql := ParseMeasurementQL(query)
		if mql.Err != nil {
			err0 = mql.Err
			return
		}
		// prepare where statement
		whereCause := make([]string, 0)
		if mql.After > 0 {
			whereCause = append(whereCause, fmt.Sprintf(`time > %d`, mql.After))
		}
		if mql.Before > 0 {
			whereCause = append(whereCause, fmt.Sprintf(`time < %d`, mql.Before))
		}
		var whereStmt string
		if len(whereCause) > 0 {
			whereStmt = fmt.Sprintf(`WHERE %s`, strings.Join(whereCause, "AND"))
		} else {
			whereStmt = ""
		}
		// prepare order by statement
		orderStmt := "DESC"
		if !mql.OrderDESC {
			orderStmt = "ASC"
		}
		// prepare total statement
		stmt := fmt.Sprintf(`SELECT %s FROM %s %s ORDER BY time %s LIMIT %d`,
			strings.Join(mql.Channel, ","), mntId, whereStmt, orderStmt, mql.Limit,
		)
		// query to rows
		rows, err := conn.Query(stmt)
		if err != nil {
			err0 = err
			return
		}
		// enumerate rows
		m := make([]Measurement, mql.Limit)
		i := 0
		numChannel := len(mql.Channel)
		for rows.Next() {
			result := make([]int64, numChannel)

			// load address of buffer into array of arguments
			addr := make([]interface{}, numChannel)
			for c := 0; c < numChannel; c++ {
				addr[c] = &result[c]
			}
			// load value into address
			rows.Scan(addr...)

			// load buffer into map
			m[i] = Measurement{}
			for c := 0; c < numChannel; c++ {
				m[i][mql.Channel[c]] = result[c]
			}
			i++
		}
		if rows.Err() != nil {
			err0 = rows.Err()
			return
		}
		ret["result"] = m[0:i]
		return
	}
}

// helper api for system status
// /sys/ip_addr: list all ip addresses of this device
// /sys/list_usb: list usb device
// /sys/list_process: list firmware backend process
func (h *APIHandler) GetSystemStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var err0 error
	ret := map[string]interface{}{}
	defer func() {
		if err0 != nil {
			ret["err"] = err0.Error()
		}
		jsonByte, _ := json.Marshal(ret)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, string(jsonByte))
	}()
	option := ps.ByName("option")
	switch option {
	case "ip_addr":
		if names, err := GetIP(); err != nil {
			err0 = err
			return
		} else {
			ret["result"] = names
		}
	case "list_process":
		status := map[string]bool{"usbshad": false, "usbsync": false}
		var err error
		for processName, _ := range status {
			if status[processName], err = IsProcessRunning(processName); err != nil {
				err0 = err
				return
			}
		}
		ret["result"] = status
	case "list_usb":
		if mapDev, err := ListUsbDevice(); err != nil {
			err0 = err
			return
		} else {
			retDev := []map[string]string{}
			for _, mdev := range mapDev {
				bus := mdev["bus_address"]
				if len(bus) > 0 && !h.StartedBusAddress[bus] {
					retDev = append(retDev, mdev)
				}
			}
			ret["result"] = retDev
		}
	case "list_service":
		status := []map[string]string{}
		var pids []string
		mapPid := map[string]bool{}

		// get current pids
		pids, err0 = ListPidFromName("usbint1")
		for _, pid := range pids {
			mapPid[pid] = true
		}
		// update each pid record; remove old
		for i, rec := range h.ServiceRecord {
			var s map[string]string

			// if record has pid
			if len(rec.Pid) > 0 {
				// if current process pid is not exists
				// stop service and remove from record
				if mapPid[rec.Pid] == false {

					StopUsbIntService(i + 1)
					StopUsbIntService(i + 1)
					h.StartedPid[rec.Pid] = false
					h.StartedBusAddress[fmt.Sprintf("%d:%d", rec.Bus, rec.Address)] = false
					h.ServiceRecord[i].Pid = ""
					h.ServiceRecord[i].PatientId = ""
					h.ServiceRecord[i].Bus = 0
					h.ServiceRecord[i].Address = 0
				}
			}
			s = map[string]string{
				"pid":     rec.Pid,
				"patient": rec.PatientId,
				"bus":     fmt.Sprintf("%d", rec.Bus),
				"address": fmt.Sprintf("%d", rec.Address),
			}
			status = append(status, s)
		}
		ret["result"] = status
	case "print_log":
		// query log file from usbapi path
		var logFile *os.File
		if logFile, err0 = os.Open(h.Conf.Log.FileName); err0 != nil {
			return
		}
		defer logFile.Close()

		buffer := make([]byte, 1000)
		var seekMove int64
		var readCount int
		if seekMove, err0 = logFile.Seek(0, 2); err0 != nil {
			return
		}
		if seekMove-1000 > 0 {
			if _, err0 = logFile.Seek(-1000, 2); err0 != nil {
				return
			}

		} else {
			if _, err0 = logFile.Seek(0, 0); err0 != nil {
				return
			}

		}
		if readCount, err0 = logFile.Read(buffer); err0 != nil {
			return
		}

		ret[h.Conf.Log.FileName] = string(buffer[0:readCount])

		// query log file from database path

		files, _ := filepath.Glob(filepath.Join(h.Conf.DB.Path, "*.log"))
		for _, file := range files {
			var logF *os.File
			if logF, err0 = os.Open(file); err0 != nil {
				return
			}
			defer logF.Close()
			buffer := make([]byte, 1000)
			var seekMove int64
			var readCount int
			if seekMove, err0 = logF.Seek(0, 2); err0 != nil {
				return
			}
			if seekMove-1000 > 0 {
				if _, err0 = logF.Seek(-1000, 2); err0 != nil {
					return
				}

			} else {
				if _, err0 = logF.Seek(0, 0); err0 != nil {
					return
				}

			}
			if readCount, err0 = logF.Read(buffer); err0 != nil {
				return
			}
			ret[file] = string(buffer[0:readCount])
		}
	case "start":
		// query string ?bus=...&addr=...&patient=...
		query := r.URL.Query()
		patient := query.Get("patient")
		busStr := query.Get("bus")
		addrStr := query.Get("addr")
		isStart := false
		if len(patient) == 0 || len(busStr) == 0 || len(addrStr) == 0 {
			err0 = fmt.Errorf("missing query")
			return
		}
		var bufInt64 int64
		if bufInt64, err0 = strconv.ParseInt(busStr, 10, 0); err0 != nil {
			return
		}
		bus := int(bufInt64)
		if bufInt64, err0 = strconv.ParseInt(addrStr, 10, 0); err0 != nil {
			return
		}
		addr := int(bufInt64)
		if len(patient) > 0 && bus != 0 && addr != 0 {
			for i, s := range h.ServiceRecord {
				if len(s.Pid) == 0 || s.Bus == 0 || s.Address == 0 {

					go func() {
						//start usb service i+1
						var err error
						if err = StopUsbIntService(i + 1); err != nil {
							log.Print(err)
						}
						if err = StartUsbIntService(i+1, patient, bus, addr); err != nil {
							log.Print(err)
							return
						}
						// get pids list from all usbint1.exe process
						var pids []string
						if pids, err = ListPidFromName("usbint1"); err != nil {
							log.Print(err)
							return
						}
						for _, pid := range pids {
							// check if not exists pid
							if val := h.StartedPid[pid]; val == false {
								h.ServiceRecord[i].Pid = pid
								h.ServiceRecord[i].PatientId = patient
								h.ServiceRecord[i].Bus = bus
								h.ServiceRecord[i].Address = addr
								h.StartedPid[pid] = true
								h.StartedBusAddress[fmt.Sprintf("%d:%d", bus, addr)] = true
							}
						}
					}()
					isStart = true
					ret["result"] = i
					break
				}
			}
			if isStart == false {
				err0 = fmt.Errorf("cannot start; no idle USB service")
			}
		} else {
			err0 = fmt.Errorf("invalid query string")
		}
	case "stop":
		// query string ?bus=...&addr=...&patient=...
		query := r.URL.Query()
		patient := query.Get("patient")
		busStr := query.Get("bus")
		addrStr := query.Get("addr")
		isStop := false
		if len(patient) == 0 || len(busStr) == 0 || len(addrStr) == 0 {
			err0 = fmt.Errorf("missing query")
			return
		}
		var bufInt64 int64
		if bufInt64, err0 = strconv.ParseInt(busStr, 10, 0); err0 != nil {
			return
		}
		bus := int(bufInt64)
		if bufInt64, err0 = strconv.ParseInt(addrStr, 10, 0); err0 != nil {
			return
		}
		addr := int(bufInt64)
		if len(patient) > 0 && bus != 0 && addr != 0 {
			for i, s := range h.ServiceRecord {
				// if exitsts pid in record; clear it
				if s.Bus == bus && s.Address == addr {
					go func() {
						if err := StopUsbIntService(i + 1); err != nil {
							log.Print(err)
						}
						if err := StopUsbIntService(i + 1); err != nil {
							log.Print(err)
							return
						}
						h.StartedPid[s.Pid] = false
						h.StartedBusAddress[fmt.Sprintf("%d:%d", bus, addr)] = false
						h.ServiceRecord[i].Pid = ""
						h.ServiceRecord[i].PatientId = ""
						h.ServiceRecord[i].Bus = 0
						h.ServiceRecord[i].Address = 0

					}()
					ret["result"] = i
					isStop = true
					break
				}
			}
			if isStop == false {
				err0 = fmt.Errorf("cannot stop; no specific USB service")
			}
		} else {
			err0 = fmt.Errorf("invalid query string")
		}
	default:
		err0 = fmt.Errorf("path %s not found", option)
	}
}

func (h *APIHandler) Close() {
	for _, conn := range h.SqliteConn {
		conn.Close()
	}
}

//------------------------------------------------------------------------------
// End APIHandler Section
//------------------------------------------------------------------------------

type MeasurementQL struct {
	Limit     int
	Channel   []string
	After     int64
	Before    int64
	OrderDESC bool
	Err       error
}

func ParseMeasurementQL(query url.Values) *MeasurementQL {
	mql := &MeasurementQL{
		Limit:     100,
		Channel:   []string{"time"},
		OrderDESC: true,
	}
	orderby := query.Get("orderby")
	limit := query.Get("limit")
	ch := query.Get("ch")
	after := query.Get("after")
	before := query.Get("before")
	if len(orderby) > 0 {
		if orderby == "asc" || orderby == "1-10" {
			mql.OrderDESC = false
		}
	}
	if len(limit) > 0 {
		if i, err := strconv.ParseInt(limit, 10, 32); err != nil {
			mql.Err = err
			return mql
		} else {
			mql.Limit = int(i)
		}
	}
	if len(ch) > 0 {
		mql.Channel = append(mql.Channel, strings.Split(ch, ",")...)
	}
	afterLength := len(after)
	if afterLength > 0 {
		var suffix byte
		if strings.HasSuffix(after, "s") {
			if afterLength == 1 {
				mql.Err = fmt.Errorf("invalid after")
				return mql
			}
			suffix = after[afterLength-2]
			if suffix == 'n' || suffix == 'm' || suffix == 'u' {
				after = after[0 : afterLength-2]
			} else {
				after = after[0 : afterLength-1]
			}
		}
		if i, err := strconv.ParseInt(after, 10, 64); err != nil {
			mql.Err = err
			return mql
		} else {
			switch suffix {
			case 'n':
			case 'm':
				i = i * 1e6
			case 'u':
				i = i * 1e3
			default:
				i = i * 1e9
			}
			mql.After = i
		}
	}
	beforeLength := len(before)
	if beforeLength > 0 {
		var suffix byte
		if strings.HasSuffix(before, "s") {
			if beforeLength == 1 {
				mql.Err = fmt.Errorf("invalid before")
				return mql
			}
			suffix = before[beforeLength-2]
			if suffix == 'n' || suffix == 'm' || suffix == 'u' {
				before = before[0 : beforeLength-2]
			} else {
				before = before[0 : beforeLength-1]
			}
		}

		if i, err := strconv.ParseInt(before, 10, 64); err != nil {
			mql.Err = err
			return mql
		} else {
			switch suffix {
			case 'n':
			case 'm':
				i = i * 1e6
			case 'u':
				i = i * 1e3
			default:
				i = i * 1e9
			}
			mql.Before = i
		}
	}
	return mql
}

// modified from:
// http://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
// and http://play.golang.org/p/BDt3qEQ_2H
func GetIP() ([]string, error) {
	ips := []string{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ips = append(ips, ip.String())
		}
	}
	if len(ips) == 0 {
		return nil, errors.New("are you connected to the network?")
	} else {
		return ips, nil
	}
}

func ListUsbDevice() (devMap []map[string]string, err error) {
	context := usb.NewContext()
	defer context.Close()
	var devices []*usb.Device

	devices, err = context.ListDevices(func(desc *usb.Descriptor) bool {
		if desc.Vendor == usb.ID(0x10c4) && desc.Product == usb.ID(0x8a40) {
			return true
		}
		return false
	})
	if err != usb.ERROR_NOT_FOUND {
		return
	}
	if len(devices) == 0 {
		err = errors.New("no devices")
	}
	for _, dev := range devices {
		defer dev.Close()

		mymap := make(map[string]string)
		if mymap["manufacturer"], err = dev.GetStringDescriptor(1); err != nil {
			return
		}
		if mymap["product"], err = dev.GetStringDescriptor(2); err != nil {
			return
		}
		mymap["vid"] = fmt.Sprintf("%X", int(dev.Vendor))
		mymap["pid"] = fmt.Sprintf("%X", int(dev.Product))
		mymap["bus_address"] = fmt.Sprintf("%d:%d", dev.Bus, dev.Address)
		devMap = append(devMap, mymap)
	}
	return
}
