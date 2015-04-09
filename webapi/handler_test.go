package webapi

import (
	"fmt"
	"github.com/nodtem66/usbint1/webapi/config"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var globalHandler = New()
var conf = config.TomlConfig{DB: config.Database{"./"}}

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
func TestWebApi_Index(t *testing.T) {
	globalHandler.Conf = &conf
	router := NewAPIRouter(globalHandler)

	router.NotFound = func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(404)
	}

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Route / failed: code %d", w.Code)
	}
	t.Log(w.Body)

	r, _ = http.NewRequest("GET", "/patient", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Route /patient failed: code %d", w.Code)
	}
	//t.Log(w.Body)
	if w.Body.String() != `{"result":["test"]}` {
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
		t.Errorf("Error for test /patient/test/tag/1")
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
		t.Errorf("Error for test /patient/test/tag/1")
	}
	t.Log(w.Body)
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
		"limit=10&ch=0&orderby=desc": MeasurementQL{Limit: 10, OrderDESC: true, Channel: 0},
		"limit=1000":                 MeasurementQL{Limit: 1000, OrderDESC: true, Channel: 0},
		"limit=1a":                   MeasurementQL{Err: err0},
		"limit=-1":                   MeasurementQL{Limit: -1},
		"orderby=asc":                MeasurementQL{Limit: 100, OrderDESC: false, Channel: 0},
		"orderby=1-10":               MeasurementQL{Limit: 100, OrderDESC: false, Channel: 0},
		"orderby=23":                 MeasurementQL{Limit: 100, OrderDESC: true, Channel: 0},
		"ch=1":                       MeasurementQL{Limit: 100, OrderDESC: true, Channel: 1},
		"ch=#12":                     MeasurementQL{Err: err0},
		"ch=-1":                      MeasurementQL{Err: err0},
		"before=10":                  MeasurementQL{Limit: 100, OrderDESC: true, Before: 10e9},
		"before=10s":                 MeasurementQL{Limit: 100, OrderDESC: true, Before: 10e9},
		"before=3ms":                 MeasurementQL{Limit: 100, OrderDESC: true, Before: 3e6},
		"before=5us":                 MeasurementQL{Limit: 100, OrderDESC: true, Before: 5e3},
		"before=9ns":                 MeasurementQL{Limit: 100, OrderDESC: true, Before: 9},
		"after=2s":                   MeasurementQL{Limit: 100, OrderDESC: true, After: 2e9},
		"after=7":                    MeasurementQL{Limit: 100, OrderDESC: true, After: 7e9},
		"after=2ms":                  MeasurementQL{Limit: 100, OrderDESC: true, After: 2e6},
		"after=2ns":                  MeasurementQL{Limit: 100, OrderDESC: true, After: 2},
		"after=2us":                  MeasurementQL{Limit: 100, OrderDESC: true, After: 2e3},
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
			actual.Channel == expect.Channel && actual.Limit == expect.Limit &&
			actual.OrderDESC == actual.OrderDESC) {
			t.Errorf("Error parse [QL: %s] [Actual %#v] [Expect %#v]", query, actual, expect)
		}
	}
}
