package util

import (
	"testing"
)

func TestFetch(t *testing.T) {
	if result, err := Fetch("http://localhost:9999"); err != nil {
		t.Error(err)
	} else {
		t.Log(result)
		metadata := result["metadata"].(map[string]interface{})
		name := metadata["name"].(string)
		t.Log(name)
	}

}

func TestGetNodes(t *testing.T) {
	if nameList, err := GetSlaveNodes(); err != nil {
		t.Error(err)
	} else {
		t.Log(nameList)
	}
}
