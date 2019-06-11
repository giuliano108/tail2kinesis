package lib

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func unmarshalString(data string, v interface{}) error {
	d := json.NewDecoder(strings.NewReader(data))
	d.UseNumber()
	if err := d.Decode(&v); err != nil {
		return err
	}
	return nil
}

func TestIdentity(t *testing.T) {
	var xform Transform
	var err error

	xform, err = NewIdentity()
	if err != nil {
		t.Errorf("%v", err)
	}

	message, err := xform.Do("something")
	if err != nil {
		t.Errorf("%v", err)
	}

	if message != "something" {
		t.Errorf("\nexpected   : %#v\ntransformed: %#v", "something", message)
	}
}

func TestAccessLog2JSON(t *testing.T) {
	var xform Transform
	var emitter string
	var err error
	var fixtureb []byte
	var transformed map[string]interface{}

	emitter, err = os.Hostname()
	if err != nil {
		log.Errorf("Could not determine hostname: %v", err)
		return
	}

	xform, err = NewAccessLog2JSON("unimplemented")
	if err == nil {
		t.Errorf("Should've returned an error")
		return
	}

	xform, err = NewAccessLog2JSON("query")

	//
	fixtureb, err = ioutil.ReadFile("testdata/access-log-query01.log")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	var expected map[string]interface{}
	var message string
	expected = map[string]interface{}{
		"_meta": map[string]interface{}{
			"time":    "08/May/2019:08:17:15 -0700",
			"path":    "/telemetry/endpoint",
			"emitter": emitter,
		},
		"uptime":     json.Number("21597960"),
		"hostname":   "thehost",
		"id":         "07f635fd2",
		"role":       "somerole",
		"event":      "end",
		"elapsed":    json.Number("252"),
		"statuscode": json.Number("2"),
	}
	message, err = xform.Do(string(fixtureb))
	if err != nil {
		t.Errorf("%v", err)
	}
	transformed = make(map[string]interface{})
	unmarshalString(message, &transformed)
	if !reflect.DeepEqual(expected, transformed) {
		t.Errorf("\nexpected   : %#v\ntransformed: %#v", expected, transformed)
	}

	//
	message, err = xform.Do("doesn't look like access log")
	if err == nil || err.Error() != "no match" {
		t.Errorf("Should've returned an error")
	}

	// no query string
	fixtureb, err = ioutil.ReadFile("testdata/access-log-query02.log")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	expected = map[string]interface{}{
		"_meta": map[string]interface{}{
			"time":    "08/May/2019:08:17:15 -0700",
			"path":    "/",
			"emitter": emitter,
		},
	}
	message, err = xform.Do(string(fixtureb))
	if err != nil {
		t.Errorf("%v", err)
	}
	transformed = make(map[string]interface{})
	unmarshalString(message, &transformed)
	if !reflect.DeepEqual(expected, transformed) {
		t.Errorf("\nexpected   : %#v\ntransformed: %#v", expected, transformed)
	}

	// invalid URL
	fixtureb, err = ioutil.ReadFile("testdata/access-log-query03.log")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	expected = map[string]interface{}{
		"_meta": map[string]interface{}{
			"time":    "08/May/2019:08:17:15 -0700",
			"path":    "",
			"emitter": emitter,
		},
	}
	message, err = xform.Do(string(fixtureb))
	if err != nil {
		t.Errorf("%v", err)
	}
	transformed = make(map[string]interface{})
	unmarshalString(message, &transformed)
	if !reflect.DeepEqual(expected, transformed) {
		t.Errorf("\nexpected   : %#v\ntransformed: %#v", expected, transformed)
	}
}
