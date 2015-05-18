package webapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/nodtem66/usbint1/config"
)

var globalHandler = New()
var conf = config.TomlConfig{DB: config.Database{"./"}, Log: config.Loginfo{FileName: "./test.log"}}

func assertEqual(t *testing.T, expect interface{}, actual interface{}) {
	if expect != actual {
		t.Errorf("[Expect: %#v] != [Actual: %#v]", expect, actual)
	}
}
func assertNotEqual(t *testing.T, expect interface{}, actual interface{}) {
	if expect == actual {
		t.Errorf("[Expect: %#v] == [Actual: %#v]", expect, actual)
	}
}
func assertChannel(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
func TestWebApi_Index(t *testing.T) {
	globalHandler.Conf = &conf
	router := NewAPIRouter(globalHandler)

	router.NotFound = func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(404)
	}

	r, _ := http.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Route / failed: code %d", w.Code)
	}

	r, _ = http.NewRequest("GET", "/patient", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Route /patient failed: code %d", w.Code)
	}
	//t.Log(w.Body)
	if w.Body.String() != `{"result":["100","test"]}` {
		t.Errorf("No Test.db inside webapi")
	}

	r, _ = http.NewRequest("GET", "/patient/test/tag", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Route /patient/test/tag failed: code %d", w.Code)
	}

	r, _ = http.NewRequest("GET", "/patient/test/tag/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Route /patient/test/tag/1 failed: code %d", w.Code)
	}

	r, _ = http.NewRequest("GET", "/patient/test/mnt", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Route /patient/test/mnt failed: code %d", w.Code)
	}

	r, _ = http.NewRequest("GET", "/patient/test/mnt/general_1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Route /patient/test/mnt/general_1 failed: code %d", w.Code)
	}

	r, _ = http.NewRequest("GET", "/invalid_url", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code == 200 {
		t.Errorf("Route /invalid_url should not be accessed")
	}
}

func TestWebApi_Tag(t *testing.T) {

	globalHandler.Conf = &conf
	router := NewAPIRouter(globalHandler)

	router.NotFound = func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(404)
	}

	r, _ := http.NewRequest("GET", "/patient/test22/tag", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Body.String() != `{"err":"no patient"}` {
		t.Errorf("Error for test /patient/test22/tag")
	}

	r, _ = http.NewRequest("GET", "/patient/test/tag/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertNotEqual(t, 200, w.Code)

	r, _ = http.NewRequest("GET", "/patient/test/tag", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[`) {
		t.Errorf("Error for test /patient/test/tag")
	}

	r, _ = http.NewRequest("GET", "/patient/100/tag", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[`) {
		t.Errorf("Error for test /patient/test/tag")
	}

	r, _ = http.NewRequest("GET", "/patient/test/tag?active", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[`) {
		t.Errorf("Error for test /patient/test/tag?active")
	}

	r, _ = http.NewRequest("GET", "/patient/test/tag?inactive", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[`) {
		t.Errorf("Error for test /patient/test/tag?inactive")
	}
}

func TestWebApi_TagId(t *testing.T) {
	globalHandler.Conf = &conf
	router := NewAPIRouter(globalHandler)
	
	router.NotFound = func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(404)
	}

	r, _ := http.NewRequest("GET", "/patient/test/tag/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":{"id":1`) {
		t.Errorf("Error for test /patient/test/tag/1: %s", w.Body.String())
	}

	r, _ = http.NewRequest("GET", "/patient/100/tag/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":{"id":1`) {
		t.Errorf("Error for test /patient/test/tag/1: %s", w.Body.String())
	}

	r, _ = http.NewRequest("GET", "/patient/test/tag/-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	assertEqual(t, `{"err":"no tag"}`, w.Body.String())

	r, _ = http.NewRequest("GET", "/patient/test/tag/10000", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	assertEqual(t, `{"err":"no tag"}`, w.Body.String())
}

func TestWebApi_Measurement(t *testing.T) {
	globalHandler.Conf = &conf
	router := NewAPIRouter(globalHandler)

	router.NotFound = func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(404)
	}
	r, _ := http.NewRequest("GET", "/patient/test/mnt", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[`) {
		t.Errorf("Error for test /patient/test/mnt")
	}
	r, _ = http.NewRequest("GET", "/patient/100/mnt", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[`) {
		t.Errorf("Error for test /patient/100/mnt")
	}
	r, _ = http.NewRequest("GET", "/patient/100/mnt/oxigen_sat_1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	t.Log(w.Body)
	if !strings.HasPrefix(w.Body.String(), `{"result":{"name":"oxigen_sat_1","`) {
		t.Errorf("Error for test /patient/100/mnt/oxigen_sat_1")
	}
	r, _ = http.NewRequest("GET", "/patient/test/mnt/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertNotEqual(t, 200, w.Code)

	r, _ = http.NewRequest("GET", "/patient/test/mnt/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if w.Body.String() != `{"err":"invalid measurement unit"}` {
		t.Errorf("Error for test /patient/test/mnt/1")
	}

	r, _ = http.NewRequest("GET", "/patient/test/mnt/general_1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":{"name":"general_1","`) {
		t.Errorf("Error for test /patient/test/mnt/general_1")
	}
	t.Log(w.Body)
}

func TestWebApi_MQL(t *testing.T) {
	err0 := fmt.Errorf("Error")
	testcase := map[string]MeasurementQL{
		"limit=10&ch=0&orderby=desc": MeasurementQL{Limit: 10, OrderDESC: true, Channel: []string{"time", "0"}},
		"limit=1000":                 MeasurementQL{Limit: 1000, OrderDESC: true, Channel: []string{"time"}},
		"limit=1a":                   MeasurementQL{Err: err0},
		"limit=-1":                   MeasurementQL{Limit: -1, OrderDESC: true, Channel: []string{"time"}},
		"orderby=asc":                MeasurementQL{Limit: 100, OrderDESC: false, Channel: []string{"time"}},
		"orderby=1-10":               MeasurementQL{Limit: 100, OrderDESC: false, Channel: []string{"time"}},
		"orderby=23":                 MeasurementQL{Limit: 100, OrderDESC: true, Channel: []string{"time"}},
		"ch=a":                       MeasurementQL{Limit: 100, OrderDESC: true, Channel: []string{"time", "a"}},
		"ch=b123-2":                  MeasurementQL{Limit: 100, OrderDESC: true, Channel: []string{"time", "b123-2"}},
		"ch=a,b,c":                   MeasurementQL{Limit: 100, OrderDESC: true, Channel: []string{"time", "a", "b", "c"}},
		"before=10":                  MeasurementQL{Limit: 100, OrderDESC: true, Before: 10e9, Channel: []string{"time"}},
		"before=10s":                 MeasurementQL{Limit: 100, OrderDESC: true, Before: 10e9, Channel: []string{"time"}},
		"before=3ms":                 MeasurementQL{Limit: 100, OrderDESC: true, Before: 3e6, Channel: []string{"time"}},
		"before=5us":                 MeasurementQL{Limit: 100, OrderDESC: true, Before: 5e3, Channel: []string{"time"}},
		"before=9ns":                 MeasurementQL{Limit: 100, OrderDESC: true, Before: 9, Channel: []string{"time"}},
		"after=2s":                   MeasurementQL{Limit: 100, OrderDESC: true, After: 2e9, Channel: []string{"time"}},
		"after=7":                    MeasurementQL{Limit: 100, OrderDESC: true, After: 7e9, Channel: []string{"time"}},
		"after=2ms":                  MeasurementQL{Limit: 100, OrderDESC: true, After: 2e6, Channel: []string{"time"}},
		"after=2ns":                  MeasurementQL{Limit: 100, OrderDESC: true, After: 2, Channel: []string{"time"}},
		"after=2us":                  MeasurementQL{Limit: 100, OrderDESC: true, After: 2e3, Channel: []string{"time"}},
		"after=2na":                  MeasurementQL{Err: err0},
		"after=2n":                   MeasurementQL{Err: err0},
		"after=n":                    MeasurementQL{Err: err0},
		"after=s":                    MeasurementQL{Err: err0},
		"after=ss":                   MeasurementQL{Err: err0},
	}
	for query, expect := range testcase {
		q, _ := url.ParseQuery(query)
		actual := ParseMeasurementQL(q)
		if actual.Err != nil && expect.Err != nil {
			continue
		} else if !(actual.After == expect.After && actual.Before == expect.Before &&
			assertChannel(actual.Channel, expect.Channel) && actual.Limit == expect.Limit &&
			actual.OrderDESC == actual.OrderDESC) {
			t.Errorf("Error parse [QL: %s] [Actual %#v] [Expect %#v]", query, actual, expect)
		}
	}
}

func TestWebApi_MeasurementQuery(t *testing.T) {
	globalHandler.Conf = &conf
	router := NewAPIRouter(globalHandler)

	router.NotFound = func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(404)
	}
	r, _ := http.NewRequest("GET", "/patient/test/mnt/general_1?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[{"time":`) {
		t.Fatal(w.Body)
	}

	r, _ = http.NewRequest("GET", "/patient/100/mnt/oxigen_sat_1?limit=1&ch=led1,led2", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[{"led1":`) {
		t.Fatal(w.Body)
	}
	t.Log(w.Body)

	r, _ = http.NewRequest("GET", "/patient/test/mnt/general_1?limit=1&ch=sync", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"result":[{"sync":0,"time":`) {
		t.Fatal(w.Body)
	}

	r, _ = http.NewRequest("GET", "/patient/test/mnt/general_1?limit=10&ch=does_not_exist_column", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `{"err":"`) {
		t.Fatal(w.Body)
	}
}

func TestWebApi_ServerFile(t *testing.T) {
	globalHandler.Conf = &conf
	router := NewAPIRouter(globalHandler)

	r, _ := http.NewRequest("GET", "/example.html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	if !strings.HasPrefix(w.Body.String(), `Hello from example.html`) {
		t.Fatal(w.Body)
	}
	r, _ = http.NewRequest("GET", "/index.html", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertNotEqual(t, 200, w.Code)
}

func TestWebApi_SystemStatus(t *testing.T) {
	globalHandler.Conf = &conf
	router := NewAPIRouter(globalHandler)

	r, _ := http.NewRequest("GET", "/sys/ip_addr", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	t.Log(w.Body)
	r, _ = http.NewRequest("GET", "/sys/list_process", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	t.Log(w.Body)
	r, _ = http.NewRequest("GET", "/sys/list_usb", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	t.Log(w.Body)
	r, _ = http.NewRequest("GET", "/sys/print_log", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assertEqual(t, 200, w.Code)
	t.Log(w.Body)
}

func TestWindow_GetIP(t *testing.T) {
	if name, err := GetIP(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(name)
	}
}

func TestWindow_IsProcessRunning(t *testing.T) {
	if isRun, err := IsProcessRunning("chrome"); err != nil {
		t.Fatal(err)
	} else {
		assertEqual(t, isRun, true)
	}
	if isRun, err := IsProcessRunning("usbint"); err != nil {
		t.Fatal(err)
	} else {
		assertEqual(t, isRun, false)
	}
}

func TestWindow_ListUSB(t *testing.T) {
	if jsonStr, err := ListUsbDevice(); err != nil {
		t.Fatal(err)
	} else {
		t.Log(jsonStr)
	}
}

func TestWindow_ListPidFromName(t *testing.T) {
	if pids, err := ListPidFromName("chrome"); err != nil {
		t.Fatal(err)
	} else {
		t.Log(pids)
	}
	if pids, err := ListPidFromName("usbint1"); err != nil {
		t.Fatal(err)
	} else {
		t.Log(pids)
	}
}

