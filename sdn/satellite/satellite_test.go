package satellite

import (
	"fmt"
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

func TestGoPackageAccuracy(t *testing.T) {
	year, month, day, hour, minute, second := 2023, 9, 18, 4, 0, 0
	constellation, err := NewConstellation("../data/test.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	for ; minute <= 30; minute++ {
		distance := constellation.Satellites[0].DistanceAtTime(
			constellation.Satellites[1], year, month, day, hour, minute, second,
		)
		fmt.Println(distance)
	}
}


