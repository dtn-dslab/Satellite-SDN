package util

import (
	"testing"
)

func TestCreateConstellation(t *testing.T) {
	constellation, err := NewConstellation("../data/beidou.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	satellite1 := constellation.Satellites[1]
	satellite2 := constellation.Satellites[2]
	satellite1.Location()
	satellite2.Location()
	t.Logf("Distance between sat1 and sat2 is %d\n", satellite1.Distance(satellite2))
}