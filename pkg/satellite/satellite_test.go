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

func TestConnection(t *testing.T) {
	constellation, err := NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	for i := 0; i < len(constellation.Satellites); i++ {
		sat1Name := constellation.Satellites[i].Name
		for j := i + 1; j < len(constellation.Satellites); j++ {
			sat2Name := constellation.Satellites[j].Name
			flag, err := constellation.isConnection(sat1Name, sat2Name)
			if err != nil {
				t.Errorf("%v\n",err)
			}
			if flag {
				t.Logf("%s and %s: %v\n", sat1Name, sat2Name, flag)
			}
		}
	}
}

func TestGenerateEdgeSet(t *testing.T) {
	constellation, err := NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	nameMap, connGraph := constellation.GenerateConnGraph()
	edgeSet := ConvertConnGraphToEdgeSet(connGraph)
	t.Logf("NameMap: %v\n", nameMap)
	t.Logf("EdgeSet: %v\n", edgeSet)
}

func TestGenerateDistanceMap(t *testing.T) {
	constellation, err := NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	nameMap, connGraph := constellation.GenerateConnGraph()
	distanceMap := constellation.GenerateDistanceMap(connGraph)
	t.Logf("NameMap: %v\n", nameMap)
	t.Logf("DistanceMap: %v\n", distanceMap)
}