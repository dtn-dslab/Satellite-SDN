package util

import (
	"reflect"
	"testing"
)

const (
	testURL1 = "http://10.0.0.11:31968/Loaction/Loaction"
	testURL2 = "http://localhost:9999/test"
)

func TestFetch(t *testing.T) {
	if result, err := Fetch(testURL1); err != nil {
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
