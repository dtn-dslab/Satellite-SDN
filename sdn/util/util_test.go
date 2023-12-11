package util

import (
	"reflect"
	"testing"
)

const (
	testURL1 = "http://10.0.0.11:32121/Location/Location"
	testURL2 = "http://localhost:30100/location"
)

func TestFetch(t *testing.T) {
	if result, err := Fetch(testURL2); err != nil {
		t.Error(err)
	} else {
		t.Log(reflect.TypeOf(result["result"]))
		t.Log(result)
	}

}

func TestGetNodes(t *testing.T) {
	if nameList, err := GetSlaveNodes(3); err != nil {
		t.Error(err)
	} else {
		t.Log(nameList)
	}
}
