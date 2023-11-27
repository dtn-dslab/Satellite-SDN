package link

import (
	"testing"

	satv1 "ws/dtn-satellite-sdn/sdn/type/v1"
	"ws/dtn-satellite-sdn/sdn/util"
)

func TestConnection(t *testing.T) {
	constellation, err := satv1.NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	for i := 0; i < len(constellation.Satellites); i++ {
		sat1Name := constellation.Satellites[i].Name
		for j := i + 1; j < len(constellation.Satellites); j++ {
			sat2Name := constellation.Satellites[j].Name
			flag, err := isConnection(constellation, sat1Name, sat2Name)
			if err != nil {
				t.Errorf("%v\n", err)
			}
			if flag {
				t.Logf("%s and %s: %v\n", sat1Name, sat2Name, flag)
			}
		}
	}
}

func TestGenerateEdgeSet(t *testing.T) {
	constellation, err := satv1.NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	connGraph := GenerateConnGraph(constellation)
	edgeSet := ConvertConnGraphToEdgeSet(connGraph)
	t.Logf("NameMap: %v\n", constellation.GetNameMap())
	t.Logf("EdgeSet: %v\n", edgeSet)
}

func TestGenerateDistanceMap(t *testing.T) {
	constellation, err := satv1.NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Errorf("%v\n", err)
	}

	connGraph := GenerateConnGraph(constellation)
	distanceMap := GenerateDistanceMap(constellation, connGraph)
	t.Logf("NameMap: %v\n", constellation.GetNameMap())
	t.Logf("DistanceMap: %v\n", distanceMap)
}

func TestGenerateIP(t *testing.T) {
	ip := util.GetVxlanIP(1, 2)
	t.Logf("IP is %s\n", ip)
	if ip != "128.16.2.1/24" {
		t.Errorf("IP Dismatch!\n")
	}
}

func TestGenerateLinkYaml(t *testing.T) {
	constellation, err := satv1.NewConstellation("../data/geodetic.txt")
	if err != nil {
		t.Error(err)
	}

	nameMap := constellation.GetNameMap()
	connGraph := GenerateConnGraph(constellation)
	edgeSet := ConvertConnGraphToEdgeSet(connGraph)
	err = LinkSyncLoop(nameMap, edgeSet, true)
	if err != nil {
		t.Error(err)
	}
}
