package satellite

import (
	"testing"
)

func TestCreateConstellation(t *testing.T) {
	constellation, err := NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	for _, satellite := range constellation.Satellites {
		long, lat, alt := satellite.Location()
		t.Logf("%s: long is %2f, lat is %2f, alt is %2f\n",
			satellite.Name, long, lat, alt)
	}
}


